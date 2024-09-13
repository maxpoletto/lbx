package mdparser

import (
	"testing"
)

func TestParseCollectionMetadata(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCollectionMetadata([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCollectionMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
