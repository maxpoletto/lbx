package metadata

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

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
	if err := checkCommonMetadata(&cm.CommonMetadata); err != nil {
		return nil, err
	}
	// Set some site-wide defaults.
	if len(cm.Filter) == 0 {
		cm.Filter = []string{"include:.*"}
	}
	if cm.SortOrder == "" {
		cm.SortOrder = "taken"
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
	if err := checkCommonMetadata(&am.CommonMetadata); err != nil {
		return nil, err
	}
	return &am, nil
}

// checkCommonMetadata checks the format of shared metadata fields.
func checkCommonMetadata(cm *CommonMetadata) error {
	// Set / check sort order.
	switch cm.SortOrder {
	case "", "taken", "taken:reverse", "name", "name:reverse", "mtime", "mtime:reverse":
		// Empty sort order is allowed to support inheritance.
		// It is set to "taken" only in the collection metadata.
	default:
		return fmt.Errorf("invalid sort order %s", cm.SortOrder)
	}
	// Check that filter entries have the form "include:FILE" or "exclude:FILE".
	for _, entry := range cm.Filter {
		parts := strings.Split(entry, ":")
		if len(parts) != 2 || (parts[0] != "include" && parts[0] != "exclude") {
			return fmt.Errorf("invalid filter entry: %s", entry)
		}
	}
	return nil
}
