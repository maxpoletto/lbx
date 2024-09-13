// mdparser implements a parser for LBX metadata in JSON format.
package mdparser

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// CommonMetadata represents metadata that applies to multiple levels
// of the LBX photo directory hierarchy.
type CommonMetadata struct {
	// Enabled is true if photo upload is enabled. Default is true.
	Enabled bool `json:"enabled"`
	// Tags is a list of tags that apply to photos.
	// In the context of an album, a tag may optionally have the format'
	// "FILENAME:TAG", where FILENAME is the filename of a photo in the album.
	Tags []string `json:"tags"`
	// SortOrder is the order in which photos are displayed. One of:
	// "name", "name:reverse", "mtime", "mtime:reverse", "taken", "taken:reverse".
	// Default is "taken".
	SortOrder string `json:"sort_order"`
	// Access is a list of credentials that are granted read access.
	Access []string `json:"access"`
	// Include is a list of filename regexps of photos to include. By default, ".*".
	Include []string `json:"include"`
	// Exclude is a list of regexps for photos to exclude. By default, "".
	Exclude []string `json:"exclude"`
}

// CollectionMetadata represents the metadata of an LBX photo collection.
type CollectionMetadata struct {
	// CommonMetadata is the metadata that applies to multiple levels.
	CommonMetadata
	// Version is the version of the metadata format.
	Version string `json:"version"`
	// Name is the name of the collection.
	Name string `json:"name"`
	// Author is the author of the collection.
	Author string `json:"author"`
	// URL is the base URL of the collection (e.g., "https://janesmith.com/photos").
	URL string `json:"url"`
	// S3AccessCode is the access code for the S3 bucket.
	S3AccessCode string `json:"s3_access_code"`
	// S3SecretKey is the secret key for the S3 bucket.
	S3SecretKey string `json:"s3_secret_key"`
	// MaxSize is the maximum photo display size (pixels of longest side), or "0" for no limit.
	MaxSize int `json:"max_size"`
}

// AlbumMetadata represents the metadata of an LBX photo album.
type AlbumMetadata struct {
	// Title is the title of the album.
	Title string `json:"title"`
	// TitlePhoto is the filename of the title photo.
	TitlePhoto string `json:"title_photo"`
	// HighlightPhoto is the filename of the highlight photo.
	HighlightPhoto string `json:"highlight_photo"`
	// Aliases is a list of path aliases for the album, relative to the
	// collection root. Aliases must be unique within the collection.
	Aliases []string `json:"aliases"`
	// Titles is a list of photo titles. The format of each entry is:
	// "FILENAME:[LANG:]TITLE". FILENAME is a photo filename, LANG is an
	// optional language code, and TITLE is the title.
	Titles []string `json:"titles"`
	// Captions is a list of photo captions. The format of each entry is:
	// "FILENAME:[LANG:]CAPTION". FILENAME is a photo filename, LANG is an
	// optional language code, and CAPTION is the caption.
	Captions []string `json:"captions"`
}

// ParseCollectionMetadata parses the metadata of an LBX photo collection.
func ParseCollectionMetadata(data []byte) (*CollectionMetadata, error) {
	urlPattern := regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+(?:\.[a-zA-Z]{2,})+`)

	var cm CollectionMetadata
	if err := json.Unmarshal(data, &cm); err != nil {
		return nil, err
	}
	// Enforce constraints and set defaults.
	// Version must be present and non-empty.
	if cm.Version == "" {
		return nil, fmt.Errorf("require metadata version = 1")
	}
	// Name must be present and non-empty.
	if cm.Name == "" {
		return nil, fmt.Errorf("require collection name")
	}
	// URL must be present and match a URL pattern.
	if cm.URL == "" {
		return nil, fmt.Errorf("require collection URL")
	} else if !urlPattern.MatchString(cm.URL) {
		return nil, fmt.Errorf("invalid collection URL")
	}
	// S3AccessCode must be present and non-empty.
	if cm.S3AccessCode == "" {
		return nil, fmt.Errorf("require S3 access code")
	}
	// S3SecretKey must be present and non-empty.
	if cm.S3SecretKey == "" {
		return nil, fmt.Errorf("require S3 secret key")
	}
	// MaxSize must be non-negative.
	if cm.MaxSize < 0 {
		return nil, fmt.Errorf("invalid max size")
	}
	// Check sort order. If nil, set default to "taken".
	switch cm.SortOrder {
	case "":
		cm.SortOrder = "taken"
	case "name", "name:reverse", "mtime", "mtime:reverse", "taken", "taken:reverse":
		// Valid sort order.
	default:
		return nil, fmt.Errorf("invalid sort order %q", cm.SortOrder)
	}
	// Check include. If nil, set default to ".*".
	if len(cm.Include) == 0 {
		cm.Include = []string{".*"}
	}

	return &cm, nil
}

// ParseAlbumMetadata parses the metadata of an LBX photo album or intermediate directory.
// If album is true, metadata corresponds to an album (media directory).
func ParseAlbumMetadata(data []byte, album bool) (*AlbumMetadata, error) {
	var am AlbumMetadata
	if err := json.Unmarshal(data, &am); err != nil {
		return nil, err
	}
	// Title must be set iff album is true.
	if !album && am.Title != "" {
		return nil, fmt.Errorf("title can only be set in an album folder")
	} else if album && am.Title == "" {
		return nil, fmt.Errorf("require album title")
	}
	// TitlePhoto, HighlightPhoto, Aliases, Titles, and Captions can only be set in an album folder.
	if !album && (am.TitlePhoto != "" || am.HighlightPhoto != "" || len(am.Aliases) > 0 || len(am.Titles) > 0 || len(am.Captions) > 0) {
		return nil, fmt.Errorf("title photo, highlight photo, aliases, titles, and captions can only be set in an album folder")
	}
	return &am, nil
}
