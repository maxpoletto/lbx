package metadata

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func containsFilename(s1 []string, s2 string) bool {
	return true
}

// ReadMetadata reads the medata of an LBX photo collection, recursively traversing
// the directory structure and parsing metadata files. Returns an error if the metadata
// cannot be read or is invalid, or otherwise a flat list of AlbumMetadata objects, one
// per album (media directory) in the collection.
func ReadMetadata(root string) ([]*AlbumMetadata, error) {
	// Read root metadata file (metadata.json) and parse it.
	fn := root + "/metadata.json"
	txt, err := os.ReadFile(fn)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %v", err)
	}
	mdCollection, err := ParseCollectionMetadata(txt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse collection metadata: %v", err)
	}

	// Recursively read metadata of subdirectories.
	mdList := []*AlbumMetadata{}
	mdAlbum := &AlbumMetadata{
		CommonMetadata: mdCollection.CommonMetadata,
	}
	dirEntries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}
	for _, e := range dirEntries {
		if !e.IsDir() {
			continue
		}
		subdir := filepath.Join(root, e.Name())
		mdl, err := recursivelyReadMetadata(subdir, mdAlbum)
		if err != nil {
			return nil, fmt.Errorf("failed to read metadata: %v", err)
		}
		mdList = append(mdList, mdl...)
	}
	// Sort the list of albums by relative path.
	for _, md := range mdList {
		md.Path, _ = filepath.Rel(root, md.Path)
	}
	sort.Slice(mdList, func(i, j int) bool {
		return mdList[i].Path < mdList[j].Path
	})
	return mdList, nil
}

// recursivelyReadMetadata reads the metadata of a directory and its subdirectories.
func recursivelyReadMetadata(path string, mdParent *AlbumMetadata) ([]*AlbumMetadata, error) {
	// Determine whether file is an album (media directory).
	// If yes, it must contain a metadata file. Parse it and return it.
	// If not, read the metadata file if it exists, then recursively read subdirectories.
	// Inherit/override attributes between parent and child metadata as appropriate.
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}
	isAlbum := true
	for _, e := range dirEntries {
		if e.IsDir() {
			isAlbum = false
			break
		}
	}
	fn := path + "/metadata.json"
	txt, err := os.ReadFile(fn)
	if err != nil {
		// Media directory must contain a metadata file.
		if isAlbum && os.IsNotExist(err) {
			return nil, fmt.Errorf("missing metadata file in media directory: %v", path)
		}
		// Non-media directory is allowed to not contain a metadata file.
		if !isAlbum && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read metadata file: %v", err)
		}
	}
	mdCur, err := ParseAlbumMetadata(txt, isAlbum)
	if err != nil {
		return nil, fmt.Errorf("failed to parse album metadata: %v", err)
	}
	// Merge metadata, implementing inheritance rules.
	mdCur.merge(mdParent)
	if isAlbum {
		mdCur.Path = path
		return []*AlbumMetadata{mdCur}, nil
	}

	mdList := []*AlbumMetadata{}
	for _, e := range dirEntries {
		if !e.IsDir() {
			continue
		}
		subdir := filepath.Join(path, e.Name())
		mdl, err := recursivelyReadMetadata(subdir, mdCur)
		if err != nil {
			return nil, fmt.Errorf("failed to read metadata: %v", err)
		}
		mdList = append(mdList, mdl...)
	}
	return mdList, nil
}
