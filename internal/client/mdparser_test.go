package mdparser

import (
	"testing"
)

func TestParseCollectionMetadata(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		want    *CollectionMetadata
	}{
		{
			name: "Valid metadata",
			input: `{
				"version": "1",
				"name": "My Collection",
				"url": "https://example.com/photos",
				"s3_access_code": "ACCESSCODE123",
				"s3_secret_key": "SECRETKEY123",
				"max_size": 1024
			}`,
			wantErr: false,
			want: &CollectionMetadata{
				Version:      "1",
				Name:         "My Collection",
				URL:          "https://example.com/photos",
				S3AccessCode: "ACCESSCODE123",
				S3SecretKey:  "SECRETKEY123",
				MaxSize:      1024,
				CommonMetadata: CommonMetadata{
					SortOrder: "taken",
					Include:   []string{".*"},
				},
			},
		},
		{
			name: "Missing version",
			input: `{
				"name": "My Collection",
				"url": "https://example.com/photos",
				"s3_access_code": "ACCESSCODE123",
				"s3_secret_key": "SECRETKEY123",
				"max_size": 1024
			}`,
			wantErr: true,
		},
		{
			name: "Invalid URL",
			input: `{
				"version": "1",
				"name": "My Collection",
				"url": "invalid-url",
				"s3_access_code": "ACCESSCODE123",
				"s3_secret_key": "SECRETKEY123",
				"max_size": 1024
			}`,
			wantErr: true,
		},
		{
			name: "Negative max size",
			input: `{
				"version": "1",
				"name": "My Collection",
				"url": "https://example.com/photos",
				"s3_access_code": "ACCESSCODE123",
				"s3_secret_key": "SECRETKEY123",
				"max_size": -1
			}`,
			wantErr: true,
		},
		{
			name: "Missing S3 access code",
			input: `{
				"version": "1",
				"name": "My Collection",
				"url": "https://example.com/photos",
				"s3_secret_key": "SECRETKEY123",
				"max_size": 1024
			}`,
			wantErr: true,
		},
		{
			name: "Missing S3 secret key",
			input: `{
				"version": "1",
				"name": "My Collection",
				"url": "https://example.com/photos",
				"s3_access_code": "ACCESSCODE123",
				"max_size": 1024
			}`,
			wantErr: true,
		},
		{
			name: "Valid sort order",
			input: `{
				"version": "1",
				"name": "My Collection",
				"url": "https://example.com/photos",
				"s3_access_code": "ACCESSCODE123",
				"s3_secret_key": "SECRETKEY123",
				"max_size": 1024,
				"sort_order": "name"
			}`,
			wantErr: false,
			want: &CollectionMetadata{
				Version:      "1",
				Name:         "My Collection",
				URL:          "https://example.com/photos",
				S3AccessCode: "ACCESSCODE123",
				S3SecretKey:  "SECRETKEY123",
				MaxSize:      1024,
				CommonMetadata: CommonMetadata{
					SortOrder: "name",
					Include:   []string{".*"},
				},
			},
		},
		{
			name: "Invalid sort order",
			input: `{
				"version": "1",
				"name": "My Collection",
				"url": "https://example.com/photos",
				"s3_access_code": "ACCESSCODE123",
				"s3_secret_key": "SECRETKEY123",
				"max_size": 1024,
				"sort_order": "invalid_sort_order"
			}`,
			wantErr: true,
		},
		{
			name: "Valid include",
			input: `{
				"version": "1",
				"name": "My Collection",
				"url": "https://example.com/photos",
				"s3_access_code": "ACCESSCODE123",
				"s3_secret_key": "SECRETKEY123",
				"max_size": 1024,
				"include": ["*.jpg", "*.png"]
			}`,
			wantErr: false,
			want: &CollectionMetadata{
				Version:      "1",
				Name:         "My Collection",
				URL:          "https://example.com/photos",
				S3AccessCode: "ACCESSCODE123",
				S3SecretKey:  "SECRETKEY123",
				MaxSize:      1024,
				CommonMetadata: CommonMetadata{
					SortOrder: "taken",
					Include:   []string{"*.jpg", "*.png"},
				},
			},
		},
		{
			name: "Empty include",
			input: `{
				"version": "1",
				"name": "My Collection",
				"url": "https://example.com/photos",
				"s3_access_code": "ACCESSCODE123",
				"s3_secret_key": "SECRETKEY123",
				"max_size": 1024,
				"include": []
			}`,
			wantErr: false,
			want: &CollectionMetadata{
				Version:      "1",
				Name:         "My Collection",
				URL:          "https://example.com/photos",
				S3AccessCode: "ACCESSCODE123",
				S3SecretKey:  "SECRETKEY123",
				MaxSize:      1024,
				CommonMetadata: CommonMetadata{
					SortOrder: "taken",
					Include:   []string{".*"},
				},
			},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCollectionMetadata([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCollectionMetadata(%v) error = %v, wantErr %v", i, err, tt.wantErr)
				return
			}
			if !tt.wantErr && !compareCollectionMetadata(got, tt.want) {
				t.Errorf("ParseCollectionMetadata(%v) = %v, want %v", i, got, tt.want)
			}
		})
	}
}

func compareCollectionMetadata(got, want *CollectionMetadata) bool {
	if got.Version != want.Version ||
		got.Name != want.Name ||
		got.URL != want.URL ||
		got.S3AccessCode != want.S3AccessCode ||
		got.S3SecretKey != want.S3SecretKey ||
		got.MaxSize != want.MaxSize ||
		got.SortOrder != want.SortOrder ||
		len(got.Include) != len(want.Include) {
		return false
	}
	for i := range got.Include {
		if got.Include[i] != want.Include[i] {
			return false
		}
	}
	return true
}

func TestParseAlbumMetadata(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		album   bool
		wantErr bool
		want    *AlbumMetadata
	}{
		{
			name: "Valid album metadata",
			input: `{
				"title": "My Album",
				"title_photo": "cover.jpg",
				"highlight_photo": "highlight.jpg",
				"aliases": ["alias1", "alias2"],
				"titles": ["photo1.jpg:Title 1", "photo2.jpg:Title 2"],
				"captions": ["photo1.jpg:Caption 1", "photo2.jpg:Caption 2"]
			}`,
			album:   true,
			wantErr: false,
			want: &AlbumMetadata{
				Title:          "My Album",
				TitlePhoto:     "cover.jpg",
				HighlightPhoto: "highlight.jpg",
				Aliases:        []string{"alias1", "alias2"},
				Titles:         []string{"photo1.jpg:Title 1", "photo2.jpg:Title 2"},
				Captions:       []string{"photo1.jpg:Caption 1", "photo2.jpg:Caption 2"},
			},
		},
		{
			name: "Missing album title",
			input: `{
				"title_photo": "cover.jpg",
				"highlight_photo": "highlight.jpg",
				"aliases": ["alias1", "alias2"],
				"titles": ["photo1.jpg:Title 1", "photo2.jpg:Title 2"],
				"captions": ["photo1.jpg:Caption 1", "photo2.jpg:Caption 2"]
			}`,
			album:   true,
			wantErr: true,
		},
		{
			name: "Non-album with title",
			input: `{
				"title": "My Album"
			}`,
			album:   false,
			wantErr: true,
		},
		{
			name: "Non-album with album-specific fields",
			input: `{
				"title_photo": "cover.jpg",
				"highlight_photo": "highlight.jpg",
				"aliases": ["alias1", "alias2"],
				"titles": ["photo1.jpg:Title 1", "photo2.jpg:Title 2"],
				"captions": ["photo1.jpg:Caption 1", "photo2.jpg:Caption 2"]
			}`,
			album:   false,
			wantErr: true,
		},
		{
			name:    "Valid non-album metadata",
			input:   `{}`,
			album:   false,
			wantErr: false,
			want:    &AlbumMetadata{},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAlbumMetadata([]byte(tt.input), tt.album)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAlbumMetadata(%v) error = %v, wantErr %v", i, err, tt.wantErr)
				return
			}
			if !tt.wantErr && !compareAlbumMetadata(got, tt.want) {
				t.Errorf("ParseAlbumMetadata(%v) = %v, want %v", i, got, tt.want)
			}
		})
	}
}

func compareAlbumMetadata(got, want *AlbumMetadata) bool {
	if got.Title != want.Title ||
		got.TitlePhoto != want.TitlePhoto ||
		got.HighlightPhoto != want.HighlightPhoto ||
		len(got.Aliases) != len(want.Aliases) ||
		len(got.Titles) != len(want.Titles) ||
		len(got.Captions) != len(want.Captions) {
		return false
	}
	for i := range got.Aliases {
		if got.Aliases[i] != want.Aliases[i] {
			return false
		}
	}
	for i := range got.Titles {
		if got.Titles[i] != want.Titles[i] {
			return false
		}
	}
	for i := range got.Captions {
		if got.Captions[i] != want.Captions[i] {
			return false
		}
	}
	return true
}
