-- SQLite schema for LBX photo metadata.

PRAGMA foreign_keys = ON;

------ Folders
CREATE TABLE folders (
    id INTEGER PRIMARY KEY ASC,
    parent_id INTEGER, -- NULL for root
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,
    FOREIGN KEY(parent_id) REFERENCES folders(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX folders_path ON folders(path);
CREATE INDEX folders_parent_id ON folders(parent_id);

------ Albums
CREATE TABLE albums (
    id INTEGER PRIMARY KEY ASC,
    folder_id INTEGER NOT NULL,
    name TEXT NOT NULL, -- Display name in URL
    path TEXT NOT NULL UNIQUE, -- Slash-separated folder path, minus name (e.g., "foo/bar/baz")
    title_photo INTEGER REFERENCES media(id),
    highlight_photo INTEGER REFERENCES media(id),
    -- sort_order: 0: name, 1: name:rev, 2:mtime, 3:mtime:rev, 4:exif_time, 5:exif_time:rev
    sort_order INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY(folder_id) REFERENCES folders(id) ON DELETE CASCADE
);
CREATE INDEX albums_folder_id ON albums(folder_id);

CREATE TABLE album_text (
    album_id INTEGER NOT NULL,
    language_code TEXT NOT NULL,
    title TEXT NOT NULL,
    blurb TEXT,
    FOREIGN KEY(album_id) REFERENCES albums(id) ON DELETE CASCADE
);
CREATE INDEX album_text_album_language ON album_text(album_id, language_code);

------ Album path aliases. Opaque strings. Must be unique.
CREATE TABLE album_aliases (
    alias TEXT PRIMARY KEY,
    album_id INTEGER NOT NULL,
    FOREIGN KEY(album_id) REFERENCES albums(id) ON DELETE CASCADE
);
CREATE INDEX album_aliases_album_id ON album_aliases(album_id);

-- Media (photos / videos).
-- Describes a logical photo, not a specific file (may have multiple resolutions).
CREATE TABLE media (
    id INTEGER PRIMARY KEY,
    album_id INTEGER NOT NULL,
    media_type INTEGER NOT NULL, -- 0: photo, 1: video
    display_name TEXT NOT NULL,
    source_filename TEXT NOT NULL,
    mtime INTEGER NOT NULL,
    -- EXIF data
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
    orientation INTEGER, -- 0: landscape, 1: portrait
    FOREIGN KEY(album_id) REFERENCES albums(id) ON DELETE CASCADE
);
CREATE INDEX media_album_id ON media(album_id);

CREATE TABLE media_text (
    media_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    caption TEXT,
    language_code TEXT NOT NULL,
    FOREIGN KEY(media_id) REFERENCES media(id) ON DELETE CASCADE
);
CREATE INDEX media_text_media_id ON media_text(media_id);

-- Blob. Describes a specific resolution of a photo or video.
CREATE TABLE blobs (
    media_id INTEGER NOT NULL,
    content_hash TEXT NOT NULL UNIQUE,
    bucket_name TEXT NOT NULL,
    object_key TEXT NOT NULL,
    height INTEGER NOT NULL,
    width INTEGER NOT NULL,
    max_dim INTEGER NOT NULL,
    FOREIGN KEY(media_id) REFERENCES media(id) ON DELETE CASCADE
);
CREATE INDEX blobs_media_id_size ON blobs(media_id, max_dim);
CREATE UNIQUE INDEX blobs_content_hash ON blobs(content_hash);

-- Tags
CREATE TABLE tags (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE album_tags (
    album_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    FOREIGN KEY(album_id) REFERENCES albums(id) ON DELETE CASCADE,
    FOREIGN KEY(tag_id) REFERENCES tags(id) ON DELETE CASCADE
);
CREATE INDEX album_tags_album_id ON album_tags(album_id);
CREATE INDEX album_tags_tag_id ON album_tags(tag_id);

CREATE TABLE media_tags (
    media_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    FOREIGN KEY(media_id) REFERENCES media(id) ON DELETE CASCADE,
    FOREIGN KEY(tag_id) REFERENCES tags(id) ON DELETE CASCADE
);
CREATE INDEX media_tags_media_id ON media_tags(media_id);
CREATE INDEX media_tags_tag_id ON media_tags(tag_id);

-- Access keys
CREATE TABLE access_keys (
    id INTEGER PRIMARY KEY,
    key TEXT NOT NULL UNIQUE,
    description TEXT
);
CREATE INDEX access_keys_key ON access_keys(key);

CREATE TABLE folder_access (
    folder_id INTEGER NOT NULL,
    access_key_id INTEGER NOT NULL,
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE CASCADE,
    FOREIGN KEY (access_key_id) REFERENCES access_keys(id) ON DELETE CASCADE
);
CREATE INDEX folder_access_folder_id ON folder_access(folder_id);
CREATE INDEX folder_access_access_key_id ON folder_access(access_key_id);

CREATE TABLE album_access (
    album_id INTEGER NOT NULL,
    access_key_id INTEGER NOT NULL,
    FOREIGN KEY (album_id) REFERENCES albums(id) ON DELETE CASCADE,
    FOREIGN KEY (access_key_id) REFERENCES access_keys(id) ON DELETE CASCADE
);
CREATE INDEX album_access_album_id ON album_access(album_id);
CREATE INDEX album_access_access_key_id ON album_access(access_key_id);

CREATE TABLE media_access (
    media_id INTEGER NOT NULL,
    access_key_id INTEGER NOT NULL,
    FOREIGN KEY (media_id) REFERENCES media(id) ON DELETE CASCADE,
    FOREIGN KEY (access_key_id) REFERENCES access_keys(id) ON DELETE CASCADE
);
CREATE INDEX media_access_media_id ON media_access(media_id);
CREATE INDEX media_access_access_key_id ON media_access(access_key_id);

------ Full-text search.
CREATE VIRTUAL TABLE album_fts USING fts4();
CREATE VIRTUAL TABLE media_fts USING fts4();
