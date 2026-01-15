package hash

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSHA256Bytes(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "empty data",
			data:     []byte{},
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "hello world",
			data:     []byte("hello world"),
			expected: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			name:     "test content",
			data:     []byte("test content\n"),
			expected: "a1fff0ffefb9eace7230c24e50731f0a91c62f9cefdfe77121c2f607125dffae",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SHA256Bytes(tt.data)
			if result != tt.expected {
				t.Errorf("SHA256Bytes(%q) = %q, want %q", tt.data, result, tt.expected)
			}
		})
	}
}

func TestSHA256File(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "empty file",
			content:  "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "hello world file",
			content:  "hello world",
			expected: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(tmpDir, tt.name+".txt")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			result, err := SHA256File(tmpFile)
			if err != nil {
				t.Fatalf("SHA256File() error = %v", err)
			}
			if result != tt.expected {
				t.Errorf("SHA256File() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSHA256FileNotExist(t *testing.T) {
	_, err := SHA256File("/nonexistent/file/path")
	if err == nil {
		t.Error("SHA256File() expected error for nonexistent file, got nil")
	}
}

func TestTruncatedHash(t *testing.T) {
	tests := []struct {
		name     string
		hash     string
		n        int
		expected string
	}{
		{
			name:     "truncate to 12 chars",
			hash:     "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
			n:        12,
			expected: "b94d27b9934d",
		},
		{
			name:     "truncate to 8 chars",
			hash:     "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
			n:        8,
			expected: "b94d27b9",
		},
		{
			name:     "n greater than hash length",
			hash:     "abc123",
			n:        100,
			expected: "abc123",
		},
		{
			name:     "n equals hash length",
			hash:     "abc123",
			n:        6,
			expected: "abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncatedHash(tt.hash, tt.n)
			if result != tt.expected {
				t.Errorf("TruncatedHash(%q, %d) = %q, want %q", tt.hash, tt.n, result, tt.expected)
			}
		})
	}
}
