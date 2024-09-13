package metadata

import (
	"reflect"
	"testing"
)

func TestMerge(t *testing.T) {
	tests := []struct {
		name     string
		receiver AlbumMetadata
		other    AlbumMetadata
		expected AlbumMetadata
	}{
		{
			name: "merge with all fields",
			receiver: AlbumMetadata{
				CommonMetadata: CommonMetadata{
					Enabled:   true,
					Tags:      []string{"tag1"},
					SortOrder: "name",
					Access:    []string{"user1"},
					Filter:    []string{"include:.*"},
				},
				Title:          "Album1",
				TitlePhoto:     "photo1.jpg",
				HighlightPhoto: "highlight1.jpg",
				Aliases:        []string{"alias1"},
				Titles:         []string{"photo1.jpg:Title1"},
				Captions:       []string{"photo1.jpg:Caption1"},
			},
			other: AlbumMetadata{
				CommonMetadata: CommonMetadata{
					Enabled:   false,
					Tags:      []string{"tag2"},
					SortOrder: "mtime",
					Access:    []string{"user2"},
					Filter:    []string{"exclude:.*"},
				},
				// These fields should never be set in a parent. Set here and below
				// to double-check they are not merged.
				Title:          "Album2",
				TitlePhoto:     "photo2.jpg",
				HighlightPhoto: "highlight2.jpg",
				Aliases:        []string{"alias2"},
				Titles:         []string{"photo2.jpg:Title2"},
				Captions:       []string{"photo2.jpg:Caption2"},
			},
			expected: AlbumMetadata{
				CommonMetadata: CommonMetadata{
					Enabled:   false,
					Tags:      []string{"tag1", "tag2"},
					SortOrder: "name",
					Access:    []string{"user1", "user2"},
					Filter:    []string{"include:.*", "exclude:.*"},
				},
				Title:          "Album1",
				TitlePhoto:     "photo1.jpg",
				HighlightPhoto: "highlight1.jpg",
				Aliases:        []string{"alias1"},
				Titles:         []string{"photo1.jpg:Title1"},
				Captions:       []string{"photo1.jpg:Caption1"},
			},
		},
		{
			name: "merge with empty receiver",
			receiver: AlbumMetadata{
				CommonMetadata: CommonMetadata{},
			},
			other: AlbumMetadata{
				CommonMetadata: CommonMetadata{
					Enabled:   true,
					Tags:      []string{"tag2"},
					SortOrder: "mtime",
					Access:    []string{"user2"},
					Filter:    []string{"exclude:.*"},
				},
				Title:          "Album2",
				TitlePhoto:     "photo2.jpg",
				HighlightPhoto: "highlight2.jpg",
				Aliases:        []string{"alias2"},
				Titles:         []string{"photo2.jpg:Title2"},
				Captions:       []string{"photo2.jpg:Caption2"},
			},
			expected: AlbumMetadata{
				CommonMetadata: CommonMetadata{
					Enabled:   false,
					Tags:      []string{"tag2"},
					SortOrder: "mtime",
					Access:    []string{"user2"},
					Filter:    []string{"exclude:.*"},
				},
				Title:          "",
				TitlePhoto:     "",
				HighlightPhoto: "",
				Aliases:        nil,
				Titles:         nil,
				Captions:       nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.receiver.merge(&tt.other)
			if !reflect.DeepEqual(tt.receiver, tt.expected) {
				t.Errorf("merge() = %v, want %v", tt.receiver, tt.expected)
			}
		})
	}
}
