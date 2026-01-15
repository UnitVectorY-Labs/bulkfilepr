package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// SHA256File computes the SHA-256 hash of a file and returns it as a hex string.
func SHA256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file for hashing: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to compute hash: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// SHA256Bytes computes the SHA-256 hash of a byte slice and returns it as a hex string.
func SHA256Bytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// TruncatedHash returns the first n characters of a SHA-256 hash.
func TruncatedHash(hash string, n int) string {
	if len(hash) <= n {
		return hash
	}
	return hash[:n]
}
