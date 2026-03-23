# LBX System Design

Based on https://docs.google.com/document/d/1Q9d3oFLbcTekokmQVq36I7wa5-BDsalr-LKfdyHTWWI/edit?tab=t.0#heading=h.vnjd78whj79n and written by Claude Opus 4.6.

## What LBX is

LBX is self-hosted software for displaying photos and videos online. It serves as a durable, low-maintenance alternative to services like Smugmug and Photobucket — a "forever home" for a personal or family media collection.

A single owner manages content from a local filesystem (the "master" copy, e.g., Lightroom exports). A CLI tool syncs that content to a server, which serves it to viewers over the web.

## Design goals

**Durability.** Easy to set up, easy to migrate, minimal ongoing maintenance. The cost of running LBX should approach the cost of storage alone. Stable, user-defined URLs that don't break when the system is moved.

**Simplicity.** Minimal dependencies. Go + SQLite + an S3-compatible store (or a local filesystem). No message queues, no caches, no background workers beyond the sync process itself.

**File-based workflow.** The owner organizes media in a directory tree and describes it with JSON manifest files. The directory tree is the source of truth. The server is a read-optimized projection of that truth.

**Beauty.** Clean, responsive viewing experience. The system gets out of the way of the photos.

**Single-user, single-instance.** One owner, one site. Multi-tenancy is explicitly out of scope — it changes every decision (auth, storage, isolation, admin) and is a separate design effort if ever needed.

---

## Architecture Overview

```
┌─────────────────────┐          ┌─────────────────────────┐
│   Master filesystem │  sync    │        LBX Server       │
│                     │ ────────▶│                         │
│  manifest.json files│  (TLS)   │  SQLite DB  +  Storage  │
│  media files        │          │  (metadata)    (blobs)  │
└─────────────────────┘          └────────────┬────────────┘
                                              │
                                              │ HTTPS
                                              ▼
                                        ┌──────────┐
                                        │  Viewers │
                                        └──────────┘
```

Three components:

1. **Master filesystem** — the owner's local directory tree. Contains media files and `manifest.json` files that describe metadata, access control, and structure. This is the canonical source of truth.
2. **`lbx` CLI** — runs on the owner's machine. Reads the master filesystem, syncs metadata and media to the server. All write operations flow through this tool.
3. **`lbxd` server** — serves content to viewers. Read-only from the viewer's perspective. Stores metadata in SQLite and media blobs in an S3-compatible store or local filesystem.

---

## Master Filesystem and Manifests

### Directory structure

The photo library is a directory tree. Each directory contains either subdirectories or media files, never both. Media directories are leaves of the tree; each leaf is an **album**.

```
photos/                          # collection root
├── manifest.json                # collection manifest (required)
├── trips/
│   ├── manifest.json            # optional: overrides for this subtree
│   ├── 2024-norway/
│   │   ├── manifest.json        # album manifest (required for albums)
│   │   ├── readme.md            # optional album description (Markdown)
│   │   ├── DSC_0001.jpg
│   │   └── DSC_0002.jpg
│   └── 2024-alps/
│       ├── manifest.json
│       └── IMG_4567.jpg
└── family/
    └── reunion-2023/
        ├── manifest.json
        └── photo1.jpg
```

### Manifest inheritance

Every manifest file is a JSON object. Fields defined in a parent directory's manifest apply to all descendants unless overridden. The collection root manifest is required; intermediate directory manifests are optional; album (leaf) manifests are required.

The inheritance rules vary by field:

| Field | Merge rule | Rationale |
|---|---|---|
| `enabled` | AND | A disabled parent disables everything below it. |
| `tags` | Union (deduplicated, sorted) | Tags accumulate downward. |
| `sort_order` | Child wins | Most-specific sort order applies. |
| `visibility` | Independent per level | Each level's visibility is its own decision (see Access Control). |
| `access` | Independent per level | Access keys are scoped to the level they're declared on. |
| `filter` | Concatenate (child rules first) | First matching rule wins. Root default `include:.*` acts as catch-all. |

Fields marked with `*` below may only appear in album manifests; all others may appear at any level.

### Collection manifest (root)

```jsonc
{
  "version": "1",
  "name": "Jane Smith's Photos",
  "author": "Jane Smith",
  "url": "https://photos.janesmith.com",
  "storage": "s3",              // "s3" or "filesystem"
  "max_size": 2048,             // max display dimension in px; 0 = no limit
  "enabled": true,
  "sort_order": "taken",
  "filter": ["include:.*"]      // default: include everything
}
```

Storage credentials (S3 keys, filesystem paths) are **not** in the manifest. They are configured via environment variables (`LBX_S3_ACCESS_KEY_ID`, `LBX_S3_SECRET_ACCESS_KEY`, `LBX_S3_BUCKET`, `LBX_S3_REGION`) or a separate `.lbx-credentials` file that is `.gitignore`d by default and never synced to the server. The server has its own credential source.

### Album manifest

```jsonc
{
  "enabled": true,
  "title": "Norway 2024",                        // * required for albums
  "title_photo": "DSC_0001.jpg",                 // * optional
  "highlight_photo": "DSC_0042.jpg",             // * optional
  "aliases": ["2024/norway-trip"],                // * optional; globally unique
  "sort_order": "taken",
  "tags": ["travel", "hiking"],
  "visibility": "public",                        // "public" or "private"
  "access": ["wedding-guests-2024"],             // keys that grant access when private
  "filter": ["exclude:raw_.*"],
  "titles": ["DSC_0001.jpg:Sunrise at Trolltunga"],           // *
  "captions": ["DSC_0001.jpg:en:Early morning on day 3"],     // *
  "captions": ["DSC_0001.jpg:it:La mattina presto al giorno 3"]
}
```

### Filter rules

Filters determine which media files in a directory are included for display. Each rule is a string of the form `include:GLOB` or `exclude:GLOB`, where GLOB is a standard glob pattern (not a regexp — globs are sufficient for filename matching and much harder to get wrong).

Rules are evaluated in order, child-first then parent. First match wins. The root manifest's default `include:.*` serves as the final catch-all.

Examples:

```jsonc
// Include everything except .jpeg files, but include bike.jpeg
["include:bike.jpeg", "exclude:*.jpeg"]
// Root default include:.* catches everything else

// Only include .jpg and .png
["exclude:*", "include:*.jpg", "include:*.png"]
```

### Naming: `manifest.json`

The canonical filename is `manifest.json` everywhere — collection root, intermediate directories, and album leaves.

---

## Access Control and Authentication

LBX has two distinct authentication concerns with different threat models and mechanisms.

### Admin authentication (owner → server)

The owner authenticates to the server for all write operations (sync, metadata updates, administration). This uses a **server API key**: a high-entropy random token shared between the CLI and the server.

- Generated during `lbx init` and stored locally (e.g., `~/.config/lbx/credentials`).
- Configured on the server via environment variable (`LBX_API_KEY`).
- Transmitted in every CLI→server request as `Authorization: Bearer <key>` over TLS.
- One key per LBX instance. Rotation is a manual operation (generate new key, update both sides).

The API key is never stored in `manifest.json`, never synced, never backed up alongside content.

### Viewer access control (viewer → server)

Viewer access is controlled at two levels: **folders** and **albums**. Each has an independent `visibility` setting: `public` or `private`.

- **Public** content is visible to anyone.
- **Private** content requires an **access key** — an opaque token presented as a URL query parameter or stored in a session cookie after first use.

The two levels compose simply:

| Folder visibility | Album visibility | Behavior |
|---|---|---|
| public | public | Anyone can browse the folder and view the album. |
| public | private | Folder listing is visible, but the album is omitted unless the viewer has the album's access key. |
| private | public | Folder listing requires the folder's access key. Once provided, all public albums within are visible. |
| private | private | Folder key required to see the listing; album key required to view the album contents. |

This is a genuine two-step concentric model: the server checks folder access first, then album access. No ambiguity, no merge logic, no three-way join.

**Per-media access control is not supported.** If a specific photo needs restricted access, put it in its own album. This eliminates the complexity of per-row ACLs on what could be hundreds of thousands of media records, and keeps the access model something a human can reason about by looking at the directory tree.

### Access keys

```sql
CREATE TABLE access_keys (
    id INTEGER PRIMARY KEY,
    token TEXT NOT NULL UNIQUE,       -- the opaque bearer token
    label TEXT,                       -- human-readable description
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    expires_at INTEGER,               -- NULL = never expires
    last_used_at INTEGER
);
```

Access keys are generated by the CLI (`lbx access create`), stored in the manifest (the `access` array), and synced to the server. They can optionally expire. The server updates `last_used_at` on use, enabling the owner to identify stale keys.

A share link looks like: `https://photos.example.com/trips/2024-norway?key=<token>`

On first use, the server sets a session cookie so the viewer doesn't need to re-present the key on every request within that album. The key in the URL is the canonical mechanism; the cookie is a UX convenience.

### Media URL privacy

Media served through the normal album viewing flow inherits the album's access check — if you can't see the album, you can't see its photos.

For direct media URLs (e.g., embedding a photo in a blog post), each media item has a system-generated **public ID**: a random, unguessable, URL-safe token (128-bit, base62-encoded). This replaces sequential integer IDs in public-facing URLs. A direct media URL looks like:

```
https://photos.example.com/_/m/<public_id>?size=L
```

This is the "unlisted YouTube video" model: security through unguessability, not per-request access checks. It is simple, CDN-friendly, and appropriate for a personal photo site. The trade-off is explicit: a leaked URL is a leaked photo. If stronger guarantees are needed in the future, the server can add a token check on this endpoint without changing the URL structure.

---

## URL Design

### Namespace separation

User content gets the clean namespace. System endpoints live under the `/_/` prefix.

```
# Viewer-facing (user content)
/trips/2024-norway                     # folder or album (resolved by path)
/trips/2024-norway/DSC_0042            # photo within album (display name)

# System endpoints
/_/api/v1/...                          # client-server sync API
/_/admin                               # admin UI (if any)
/_/login                               # login flow
/_/logout                              # logout

# Direct media access
/_/m/<public_id>                       # media by public ID
/_/m/<public_id>?size=L                # with size hint
/_/m/<public_id>?size=L&fmt=webp       # with format hint
/_/c/<content_hash>                    # media by content hash (stable across renames)
```

This avoids any collision between user-created folder names and system paths. A folder named `admin` or `api` works fine because system paths are prefixed.

### Path resolution

When the server receives a request for `/<path>`:

1. Look up `<path>` as a folder path. If found, serve the folder listing.
2. Look up `<path>` as an album path. If found, serve the album.
3. Look up `<path>` as an album alias. If found, redirect or serve the aliased album.
4. If `<path>` has the form `<prefix>/<leaf>`, look up `<prefix>` as an album (or alias) and `<leaf>` as a media display name within that album. If found, serve the photo page.
5. 404.

Real paths always take priority over aliases (step 2 before step 3). This is the simplest resolution model and prevents aliases from shadowing real content.

### Size and format

Representation variants (size, format) are specified as query parameters, not path segments:

| Parameter | Values | Default |
|---|---|---|
| `size` | `thumb`, `small`, `medium`, `large`, `original` | `large` |
| `fmt` | `jpeg`, `webp`, `avif` | server decides based on `Accept` header |

Named sizes map to maximum pixel dimensions configured at the site level. Using named sizes rather than arbitrary pixel values keeps the set of generated variants bounded and cacheable.

---

## Alias System

Aliases provide stable, user-friendly alternative paths to albums. They are defined in album manifests and must be globally unique across the site.

### Rules

1. **Aliases are album-only.** Folder aliases are not supported. If needed in the future, a `folder_aliases` table with identical structure can be added. Omitting folder aliases now keeps the feature simple and well-defined.
2. **Aliases are terminal.** An alias resolves to exactly one album. Appending further path segments to an alias (e.g., `/my-alias/photo-name`) resolves via the standard "album + media name" logic in step 4 of path resolution.
3. **Real paths take priority.** If an alias collides with a real folder or album path, the real path wins. The alias is effectively shadowed. The `lbx sync` command warns when this happens.
4. **Uniqueness is enforced at sync time.** Duplicate aliases across albums are rejected. The error surfaces during `lbx sync`, not silently at serving time.

### Schema

```sql
CREATE TABLE album_aliases (
    alias TEXT PRIMARY KEY,           -- the alias path, e.g., "2024/norway-trip"
    album_id INTEGER NOT NULL,
    FOREIGN KEY(album_id) REFERENCES albums(id) ON DELETE CASCADE
);
```

---

## Database Schema

SQLite. Single file. Backed up by copying the file (or using SQLite's backup API).

### Site (singleton)

```sql
CREATE TABLE site (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    name TEXT NOT NULL,
    author TEXT,
    url TEXT NOT NULL,
    max_size INTEGER NOT NULL DEFAULT 0,
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);
```

### Folders

```sql
CREATE TABLE folders (
    id INTEGER PRIMARY KEY,
    parent_id INTEGER,                -- NULL for root
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,        -- full path from root, e.g., "trips/2024"
    visibility TEXT NOT NULL DEFAULT 'public' CHECK (visibility IN ('public', 'private')),
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
    FOREIGN KEY(parent_id) REFERENCES folders(id) ON DELETE CASCADE
);
CREATE INDEX folders_parent_id ON folders(parent_id);
```

### Albums

```sql
CREATE TABLE albums (
    id INTEGER PRIMARY KEY,
    folder_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,        -- full path, e.g., "trips/2024/norway"
    visibility TEXT NOT NULL DEFAULT 'public' CHECK (visibility IN ('public', 'private')),
    enabled INTEGER NOT NULL DEFAULT 1,
    title_photo_id INTEGER,
    highlight_photo_id INTEGER,
    sort_order INTEGER NOT NULL DEFAULT 4,
        -- 0:name, 1:name:rev, 2:mtime, 3:mtime:rev, 4:exif_time, 5:exif_time:rev
    manifest_hash TEXT,               -- content hash of the album's manifest.json
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
    FOREIGN KEY(folder_id) REFERENCES folders(id) ON DELETE CASCADE,
    FOREIGN KEY(title_photo_id) REFERENCES media(id),
    FOREIGN KEY(highlight_photo_id) REFERENCES media(id)
);
CREATE INDEX albums_folder_id ON albums(folder_id);
```

### Album text (i18n)

```sql
CREATE TABLE album_text (
    album_id INTEGER NOT NULL,
    lang TEXT NOT NULL DEFAULT 'en',
    title TEXT NOT NULL,
    blurb TEXT,                        -- short text; long-form goes in readme.md
    PRIMARY KEY(album_id, lang),
    FOREIGN KEY(album_id) REFERENCES albums(id) ON DELETE CASCADE
);
```

### Media

```sql
CREATE TABLE media (
    id INTEGER PRIMARY KEY,
    public_id TEXT NOT NULL UNIQUE,    -- random unguessable token for public URLs
    album_id INTEGER NOT NULL,
    media_type INTEGER NOT NULL CHECK (media_type IN (0, 1)),  -- 0:photo, 1:video
    display_name TEXT NOT NULL,        -- URL-friendly name derived from filename
    source_filename TEXT NOT NULL,     -- original filename on master filesystem
    mtime INTEGER NOT NULL,           -- filesystem mtime at sync time
    -- EXIF
    exif_time INTEGER,
    latitude REAL,
    longitude REAL,
    camera TEXT,
    lens TEXT,
    focal_length REAL,
    exposure_time REAL,
    aperture REAL,
    iso INTEGER,
    flash INTEGER,
    orientation INTEGER,              -- 0:landscape, 1:portrait
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
    FOREIGN KEY(album_id) REFERENCES albums(id) ON DELETE CASCADE
);
CREATE INDEX media_album_id ON media(album_id);
```

### Media text (i18n)

```sql
CREATE TABLE media_text (
    media_id INTEGER NOT NULL,
    lang TEXT NOT NULL DEFAULT 'en',
    title TEXT,
    caption TEXT,
    PRIMARY KEY(media_id, lang),
    FOREIGN KEY(media_id) REFERENCES media(id) ON DELETE CASCADE
);
```

### Blobs (storage-agnostic)

```sql
CREATE TABLE blobs (
    id INTEGER PRIMARY KEY,
    media_id INTEGER NOT NULL,
    content_hash TEXT NOT NULL UNIQUE,
    storage_path TEXT NOT NULL,        -- "s3://bucket/key" or relative local path
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    max_dim INTEGER NOT NULL,          -- max(width, height) for size-based lookups
    byte_size INTEGER NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    FOREIGN KEY(media_id) REFERENCES media(id) ON DELETE CASCADE
);
CREATE INDEX blobs_media_id_dim ON blobs(media_id, max_dim);
```

The `storage_path` is an opaque string interpreted by the configured storage backend. For S3: `s3://bucket-name/object-key`. For local filesystem: a relative path from the configured media root. Migrating between storage types is a matter of rewriting `storage_path` values and moving the files — no schema change needed.

### Tags

```sql
CREATE TABLE tags (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE album_tags (
    album_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY(album_id, tag_id),
    FOREIGN KEY(album_id) REFERENCES albums(id) ON DELETE CASCADE,
    FOREIGN KEY(tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

CREATE TABLE media_tags (
    media_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY(media_id, tag_id),
    FOREIGN KEY(media_id) REFERENCES media(id) ON DELETE CASCADE,
    FOREIGN KEY(tag_id) REFERENCES tags(id) ON DELETE CASCADE
);
```

### Access control

```sql
CREATE TABLE access_keys (
    id INTEGER PRIMARY KEY,
    token TEXT NOT NULL UNIQUE,
    label TEXT,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    expires_at INTEGER,
    last_used_at INTEGER
);

CREATE TABLE folder_access (
    folder_id INTEGER NOT NULL,
    access_key_id INTEGER NOT NULL,
    PRIMARY KEY(folder_id, access_key_id),
    FOREIGN KEY(folder_id) REFERENCES folders(id) ON DELETE CASCADE,
    FOREIGN KEY(access_key_id) REFERENCES access_keys(id) ON DELETE CASCADE
);

CREATE TABLE album_access (
    album_id INTEGER NOT NULL,
    access_key_id INTEGER NOT NULL,
    PRIMARY KEY(album_id, access_key_id),
    FOREIGN KEY(album_id) REFERENCES albums(id) ON DELETE CASCADE,
    FOREIGN KEY(access_key_id) REFERENCES access_keys(id) ON DELETE CASCADE
);
```

No `media_access` table. Per-media ACLs are not supported (see Access Control section).

### Full-text search

```sql
CREATE VIRTUAL TABLE album_fts USING fts5(
    title, blurb, tags,
    content='album_text',
    content_rowid='rowid'
);

CREATE VIRTUAL TABLE media_fts USING fts5(
    title, caption, camera, lens,
    content='media_text',
    content_rowid='rowid'
);
```

FTS5 (not FTS4) with external content tables to avoid duplicating data. Triggers on the source tables keep the FTS index in sync.

### View statistics

```sql
CREATE TABLE view_log (
    id INTEGER PRIMARY KEY,
    entity_type TEXT NOT NULL CHECK (entity_type IN ('folder', 'album', 'media')),
    entity_id INTEGER NOT NULL,
    viewed_at INTEGER NOT NULL DEFAULT (unixepoch()),
    -- No PII. No IP addresses. Just counts over time.
    -- Aggregate periodically and prune raw rows.
);
CREATE INDEX view_log_entity ON view_log(entity_type, entity_id, viewed_at);
```

---

## Client-Server Sync Protocol

### Overview

The sync protocol is designed around the principle that the master filesystem is the source of truth. The server never modifies its own state except in response to a sync operation from the CLI.

All sync requests are authenticated with the admin API key (`Authorization: Bearer <key>`) over TLS.

### Sync negotiation

When `lbx sync` runs:

1. **Client** computes a content hash of each `manifest.json` in the tree.
2. **Client** sends the set of `(directory_path, manifest_hash)` pairs to the server.
3. **Server** compares against stored `manifest_hash` values (on `albums` and a corresponding field on `folders`). Returns the list of paths where hashes differ or are missing.
4. **Client** sends full manifests and media file lists for changed directories.
5. For each changed directory, **client** sends `(filename, content_hash, mtime)` for each media file. **Server** returns the subset of files it doesn't already have (by content hash).
6. **Client** uploads missing media files. Each upload is individually transactional (the file either fully lands or doesn't). Album-level sync is not atomic, to allow progress on slow connections.
7. **Client** sends updated metadata for changed albums/folders.
8. **Server** acknowledges completion and returns a sync counter.

The client stores the sync counter locally. On subsequent syncs, it can send the counter to the server, which responds with "everything changed since counter N" — enabling fast incremental syncs without re-hashing the entire tree.

### Deletions

When a directory or media file is removed from the master filesystem:
- The client detects its absence during tree traversal.
- The client explicitly tells the server to delete it.
- The server soft-deletes (or hard-deletes with CASCADE) the corresponding records.
- Blob storage cleanup (deleting actual media files from S3/disk) can happen asynchronously.

### API endpoints

```
POST   /_/api/v1/sync/negotiate     # exchange manifest hashes
POST   /_/api/v1/sync/media-check   # exchange file content hashes for a directory
PUT    /_/api/v1/media               # upload a media file
PUT    /_/api/v1/metadata            # update album/folder/site metadata
DELETE /_/api/v1/media/:public_id    # delete a media file
DELETE /_/api/v1/album/:path         # delete an album
DELETE /_/api/v1/folder/:path        # delete a folder
GET    /_/api/v1/status              # server status, stats, last sync time
```

All request/response bodies are JSON except media uploads (multipart form data).

---

## CLI

The CLI is the owner's interface to LBX. It reads the master filesystem and communicates with the server.

```
lbx init                              # interactive setup; generates API key, creates root manifest
lbx doctor                            # sanity check local + remote state
lbx status [--verbose]                # show sync status and stats
lbx sync [--dry-run] [--verbose]      # sync everything
lbx sync <path> [--dry-run]           # sync a specific subtree

lbx album list                        # list all albums with photo counts and view stats
lbx album status [<path>]             # show album details
lbx album read [<path>] [<key>]       # show effective (inherited) metadata for an album
lbx album set <path> <key> <value>    # update manifest.json and sync
lbx album enable <path>               # shortcut for set enabled true
lbx album disable <path>              # shortcut for set enabled false

lbx alias add <album-path> <alias>    # add an alias
lbx alias remove <alias>              # remove an alias
lbx alias list                        # list all aliases

lbx access create [--label <text>] [--expires <duration>]   # create an access key
lbx access list                       # list all access keys with last-used timestamps
lbx access revoke <token>             # revoke an access key
```

All mutating commands sync immediately by default. `--no-sync` defers the sync.

---

## Media Processing

On upload, the server generates multiple resolutions of each media file:

| Name | Max dimension | Use case |
|---|---|---|
| `thumb` | 300px | Album grid thumbnails |
| `small` | 800px | Mobile display |
| `medium` | 1600px | Desktop display |
| `large` | 2400px | Full-screen / high-DPI |
| `original` | as-uploaded | Download (subject to `max_size` limit) |

If the site's `max_size` is set, the `original` variant is capped at that dimension. The raw uploaded file is stored but never served beyond the configured limit.

EXIF data is extracted during processing and stored in the `media` table. EXIF orientation is applied to generated variants (they are always stored in display orientation).

### Video

Video support is limited to storage and thumbnail extraction in v1. Full video transcoding (HLS/DASH, multiple qualities) is deferred. The schema supports `media_type = 1` (video) and the blob table can store video variants, but the processing pipeline, playback UI, and video-specific metadata (duration, codec) are future work.

---

## Search-Engine Privacy

All server responses include `X-Robots-Tag: noindex, nofollow`. The server serves a `robots.txt` at the site root that disallows all crawlers. HTML pages include `<meta name="robots" content="noindex">`. This is not configurable — LBX is for private collections, not public galleries.

---

## Open Questions

1. **Album description: manifest field or readme.md?** Currently the design supports both `album_text.blurb` in the database (synced from manifest) and a `readme.md` file (Markdown, richer formatting). Do we need both, or should `readme.md` be the sole mechanism for album descriptions?

2. **Content-hash URLs.** The `/_/c/<content_hash>` endpoint provides stable URLs across renames and re-organizations. Should these be access-controlled (check the album the media belongs to) or unguessable-only (the hash itself is the credential)? Content hashes are deterministic, so they're not unguessable in the same way `public_id` tokens are.

3. **Tag hierarchy.** Tags are currently flat strings. Is hierarchical tagging (e.g., `travel/europe/norway`) needed? Flat tags with conventions (`travel-europe-norway`) are simpler and can be promoted to hierarchical later if needed.

4. **Folder aliases.** Explicitly deferred. If the use case for cross-referencing subtrees (e.g., `family/hikes/2024-norway` and `2024/norway-trip` where both are folders containing multiple albums) proves common, folder aliases can be added with an identical mechanism to album aliases.
