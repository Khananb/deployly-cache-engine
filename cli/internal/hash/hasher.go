package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// GenerateCacheKey creates a unique string from a prefix and a file's SHA-256 hash
func GenerateCacheKey(prefix string, lockfilePath string) (string, error) {
	file, err := os.Open(lockfilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open lockfile: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to hash lockfile: %w", err)
	}

	checksum := hex.EncodeToString(hash.Sum(nil))
	
	if prefix == "" {
		return checksum, nil
	}

	return fmt.Sprintf("%s-%s", prefix, checksum), nil
}
