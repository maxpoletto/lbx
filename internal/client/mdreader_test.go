package metadata

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadMetadata(t *testing.T) {
	rootDir := createTempDir(t)
	defer os.RemoveAll(rootDir)

	initCollection(t, rootDir)
	initAlbum(t, rootDir, "subdir", `{
		"version": "1",
		"enabled": true,
		"title": "Test Album",
		"tags": ["tag1", "tag2"],
		"sort_order": "taken",
		"access": ["user1", "user2"],
		"filter": ["include:.*"]
	}`)
	mdList, err := ReadMetadata(rootDir)
	if err != nil {
		t.Fatalf("ReadMetadata failed: %v", err)
	}

	if len(mdList) != 1 {
		t.Fatalf("Expected 1 album metadata, got %d", len(mdList))
	}
	if !mdList[0].CommonMetadata.Enabled {
		t.Fatalf("Expected Enabled to be true, got false")
	}
	if len(mdList[0].CommonMetadata.Tags) != 2 {
		t.Fatalf("Expected 2 tags, got %d", len(mdList[0].CommonMetadata.Tags))
	}
}

func TestCollectionNoMetadata(t *testing.T) {
	rootDir := createTempDir(t)
	defer os.RemoveAll(rootDir)
	_, err := ReadMetadata(rootDir)
	if err == nil {
		t.Fatalf("Expected error for missing metadata file, got nil")
	}
}

func TestAlbumNoMetadata(t *testing.T) {
	rootDir := createTempDir(t)
	defer os.RemoveAll(rootDir)

	initCollection(t, rootDir)
	subDir := filepath.Join(rootDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}
	createFile(t, subDir, "foo.jpg", "")
	_, err = ReadMetadata(rootDir)
	if err == nil {
		t.Fatalf("Expected error for missing metadata file, got nil")
	}
}

func TestReadBadIntermediateMetadata(t *testing.T) {
	rootDir := createTempDir(t)
	defer os.RemoveAll(rootDir)

	initCollection(t, rootDir)
	initAlbum(t, rootDir, "subdir1", `{
		"version": "1",
		"enabled": true,
		"title": "Intermediate directory should not have title"
		"tags": ["tag1", "tag2"],
		"sort_order": "taken",
		"access": ["user1", "user2"],
		"filter": ["include:.*"],
	}`)
	initAlbum(t, rootDir, "subdir1/album1", `{
		"version": "1",
		"enabled": true,
		"title": "Test Album",
		"tags": ["tag1", "tag2"],
		"sort_order": "taken",
		"access": ["user1", "user2"],
		"filter": ["include:.*"],
	}`)
	_, err := ReadMetadata(rootDir)
	if err == nil {
		t.Fatalf("Expected error for malformed intermediate metadata, got nil")
	}
}

func TestMultipleAlbums(t *testing.T) {
	rootDir := createTempDir(t)
	defer os.RemoveAll(rootDir)

	initCollection(t, rootDir)
	initAlbum(t, rootDir, "subdir1", `{
		"version": "1",
		"enabled": true,
		"tags": ["tag1", "tag2"],
		"sort_order": "mtime",
		"access": ["user1", "user2"],
		"filter": ["include:.*"]
	}`)
	initAlbum(t, rootDir, "subdir1/album1", `{
		"version": "1",
		"enabled": true,
		"title": "Test Album1",
		"tags": ["tag1", "tag3"],
		"access": ["user1", "user2"],
		"filter": ["exclude:.*\\.png"]
	}`)
	initAlbum(t, rootDir, "subdir1/album2", `{
		"version": "1",
		"enabled": true,
		"title": "Test Album2",
		"tags": ["tag3", "tag4"],
		"sort_order": "taken",
		"access": ["user1", "user2", "user3"],
		"filter": ["include:.*"]
	}`)
	initAlbum(t, rootDir, "album3", `{
		"version": "1",
		"enabled": false,
		"title": "Test Album3",
		"tags": ["tag1", "tag2"],
		"sort_order": "taken:reverse",
		"access": ["user1", "user2"]
	}`)

	mdList, err := ReadMetadata(rootDir)
	if err != nil {
		t.Fatalf("ReadMetadata failed: %v", err)
	}

	if len(mdList) != 3 {
		t.Fatalf("Expected 3 album metadata, got %d", len(mdList))
	}
	m0, m1, m2 := mdList[0], mdList[1], mdList[2]

	if m0.Path != "album3" {
		t.Fatalf("Expected album3, got %s", mdList[0].Path)
	}
	if m1.Path != "subdir1/album1" {
		t.Fatalf("Expected subdir1/album1, got %s", mdList[1].Path)
	}
	if m2.Path != "subdir1/album2" {
		t.Fatalf("Expected subdir1/album2, got %s", mdList[2].Path)
	}
	if (len(m0.CommonMetadata.Tags) != 2) || (len(m1.CommonMetadata.Tags) != 3) || (len(m2.CommonMetadata.Tags) != 4) {
		t.Fatalf("Expected 2, 3, 4 tags, got %d, %d, %d", len(m0.CommonMetadata.Tags), len(m1.CommonMetadata.Tags), len(m2.CommonMetadata.Tags))
	}
	if m0.CommonMetadata.SortOrder != "taken:reverse" || m1.CommonMetadata.SortOrder != "mtime" || m2.CommonMetadata.SortOrder != "taken" {
		t.Fatalf("Expected sort order taken:reverse, mtime, taken, got %s, %s, %s", m0.CommonMetadata.SortOrder, m1.CommonMetadata.SortOrder, m2.CommonMetadata.SortOrder)
	}
	if m0.CommonMetadata.Enabled || !m1.CommonMetadata.Enabled || !m2.CommonMetadata.Enabled {
		t.Fatalf("Expected enabled false, true, true, got %v, %v, %v", m0.CommonMetadata.Enabled, m1.CommonMetadata.Enabled, m2.CommonMetadata.Enabled)
	}
	if len(m0.CommonMetadata.Access) != 2 || len(m1.CommonMetadata.Access) != 2 || len(m2.CommonMetadata.Access) != 3 {
		t.Fatalf("Expected 2, 2, 3 access, got %d, %d, %d", len(m0.CommonMetadata.Access), len(m1.CommonMetadata.Access), len(m2.CommonMetadata.Access))
	}
	if len(m0.CommonMetadata.Filter) != 1 || len(m1.CommonMetadata.Filter) != 3 || len(m2.CommonMetadata.Filter) != 3 {
		t.Fatalf("Expected 1, 3, 3 filter, got %v, %v, %v", m0.CommonMetadata.Filter, m1.CommonMetadata.Filter, m2.CommonMetadata.Filter)
	}
	if m1.CommonMetadata.Filter[0] != "exclude:.*\\.png" || m1.CommonMetadata.Filter[1] != "include:.*" || m1.CommonMetadata.Filter[2] != "include:.*" {
		t.Fatalf("Expected [exclude:.*\\.png, include:.*], got %s", m1.CommonMetadata.Filter)
	}
}

func createTempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "metadata_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	return dir
}

func createFile(t *testing.T, dir, name, content string) {
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create file %s: %v", path, err)
	}
}

func initCollection(t *testing.T, rootDir string) {
	// Create a valid collection metadata file.
	mdCollection := `{
			"version": "1",
			"enabled": true,
			"name": "Test Collection",
			"url": "https://example.com",
			"s3_access_code": "access",
			"s3_secret_key": "secret",
			"tags": ["tag1", "tag2"],
			"sort_order": "taken",
			"access": ["user1", "user2"],
			"filter": ["include:.*"]
		}`
	createFile(t, rootDir, "metadata.json", mdCollection)
}

func initAlbum(t *testing.T, rootDir, path, md string) {
	subDir := filepath.Join(rootDir, path)
	err := os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}
	createFile(t, subDir, "metadata.json", md)
}
