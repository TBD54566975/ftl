package buildengine

import (
	"bytes"
	"crypto/sha256"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

// CompareFileHashes compares the hashes of the files in the oldFiles and newFiles maps.
//
// Returns true if the hashes are equal, false otherwise.
//
// If false, the returned string will be the file that caused the difference,
// prefixed with "+" if it's a new file, or "-" if it's a removed file.
func CompareFileHashes(oldFiles, newFiles map[string][]byte) (string, bool) {
	for key, hash1 := range oldFiles {
		hash2, exists := newFiles[key]
		if !exists {
			return "-" + key, false
		}
		if !bytes.Equal(hash1, hash2) {
			return key, false
		}
	}

	for key := range newFiles {
		if _, exists := oldFiles[key]; !exists {
			return "+" + key, false
		}
	}

	return "", true
}

// ComputeFileHashes computes the SHA256 hash of all (non-git-ignored) files in
// the given directory.
func ComputeFileHashes(dir string) (map[string][]byte, error) {
	config, err := LoadModuleConfig(dir)
	if err != nil {
		return nil, err
	}

	fileHashes := make(map[string][]byte)
	err = WalkDir(dir, func(srcPath string, entry fs.DirEntry) error {
		for _, pattern := range config.Watch {
			relativePath, err := filepath.Rel(dir, srcPath)
			if err != nil {
				return err
			}

			match, err := doublestar.PathMatch(pattern, relativePath)
			if err != nil {
				return err
			}

			if match && !entry.IsDir() {
				file, err := os.Open(srcPath)
				if err != nil {
					return err
				}

				hasher := sha256.New()
				if _, err := io.Copy(hasher, file); err != nil {
					_ = file.Close()
					return err
				}

				fileHashes[srcPath] = hasher.Sum(nil)

				if err := file.Close(); err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return fileHashes, err
}
