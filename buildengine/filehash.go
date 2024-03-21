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

type FileChangeType rune

func (f FileChangeType) String() string { return string(f) }
func (f FileChangeType) GoString() string {
	switch f {
	case FileAdded:
		return "buildengine.FileAdded"
	case FileRemoved:
		return "buildengine.FileRemoved"
	case FileChanged:
		return "buildengine.FileChanged"
	default:
		panic("unknown file change type")
	}
}

const (
	FileAdded   FileChangeType = '+'
	FileRemoved FileChangeType = '-'
	FileChanged FileChangeType = '*'
)

type FileHashes map[string][]byte

// CompareFileHashes compares the hashes of the files in the oldFiles and newFiles maps.
//
// Returns true if the hashes are equal, false otherwise.
//
// If false, the returned string will be a file that caused the difference and the
// returned FileChangeType will be the type of change that occurred.
func CompareFileHashes(oldFiles, newFiles FileHashes) (FileChangeType, string, bool) {
	for key, hash1 := range oldFiles {
		hash2, exists := newFiles[key]
		if !exists {
			return FileRemoved, key, false
		}
		if !bytes.Equal(hash1, hash2) {
			return FileChanged, key, false
		}
	}

	for key := range newFiles {
		if _, exists := oldFiles[key]; !exists {
			return FileAdded, key, false
		}
	}

	return ' ', "", true
}

// ComputeFileHashes computes the SHA256 hash of all (non-git-ignored) files in
// the given directory.
func ComputeFileHashes(module Project) (FileHashes, error) {
	config := module.Config()

	fileHashes := make(FileHashes)
	err := WalkDir(config.Dir, func(srcPath string, entry fs.DirEntry) error {
		for _, pattern := range config.Watch {
			relativePath, err := filepath.Rel(config.Dir, srcPath)
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
