// metadata represents metadata for an LBX photo collection.
package metadata

import (
	"sort"
)

// CommonMetadata represents metadata that applies to multiple levels
// of the LBX photo directory hierarchy.
type CommonMetadata struct {
	// Enabled is true if photo upload is enabled. Default is false.
	// If a parent is disabled, all children are disabled.
	Enabled bool `json:"enabled"`
	// Tags is a list of tags that apply to photos.
	// In the context of an album, a tag may optionally have the format'
	// "FILENAME:TAG", where FILENAME is the filename of a photo in the album.
	// Tags accumulate from parent to child.
	Tags []string `json:"tags"`
	// SortOrder is the order in which photos are displayed. One of:
	// "name", "name:reverse", "mtime", "mtime:reverse", "taken", "taken:reverse".
	// Default is "taken". Child sort order overrides parent sort order.
	SortOrder string `json:"sort_order"`
	// Access is a list of credentials that are granted read access.
	// Access accumulates from parent to child. An empty access list means public access.
	Access []string `json:"access"`
	// Filter is an ordered list of filters to apply to photos to determine which ones
	// are uploaded for display by LBX.
	// Each entry has the form: "include:FILE" or "exclude:FILE". FILE is a filename or a regexp.
	// Default is ["include:.*"].
	// Filters are evaluated sequentially starting with the album directory and moving out.
	// First rule to match wins. If no rule matches, photo is not included.
	Filter []string `json:"filter"`
}

// CollectionMetadata represents the metadata of an LBX photo collection.
type CollectionMetadata struct {
	// CommonMetadata is the metadata that applies to collections or albums.
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
	// CommonMetadata is the metadata that applies to collections or albums.
	CommonMetadata
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
	// Path is the path of the album relative to the collection root.
	Path string
}

// merge merges the receiver metadata with the given metadata. The receiver
// is downstream (further nested in the directory hierarchy) from the given metadata.
func (m *AlbumMetadata) merge(other *AlbumMetadata) {
	m.Enabled = m.Enabled && other.Enabled
	m.Tags = mergeLists(m.Tags, other.Tags)
	if m.SortOrder == "" {
		m.SortOrder = other.SortOrder
	}
	m.Access = mergeLists(m.Access, other.Access)
	m.Filter = append(m.Filter, other.Filter...)
}

// mergeLists merges two lists of strings, unifying any duplicates and sorting the result.
func mergeLists(a, b []string) []string {
	// Merge access credentials, unifying any duplicates.
	m := make(map[string]bool)
	for _, a := range a {
		m[a] = true
	}
	for _, a := range b {
		m[a] = true
	}
	l := []string{}
	for a := range m {
		l = append(l, a)
	}
	sort.Strings(l)
	return l
}
