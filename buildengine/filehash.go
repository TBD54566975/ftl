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
func ComputeFileHashes(module Module) (FileHashes, error) {
	config := module.Config

	fileHashes := make(FileHashes)
	rootDirs := computeRootDirs(config.Dir, config.Watch)

	for _, rootDir := range rootDirs {
		err := WalkDir(rootDir, func(srcPath string, entry fs.DirEntry) error {
			if entry.IsDir() {
				return nil
			}
			hash, matched, err := ComputeFileHash(rootDir, srcPath, config.Watch)
			if err != nil {
				return err
			}
			if !matched {
				return nil
			}
			fileHashes[srcPath] = hash
			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return fileHashes, nil
}

func ComputeFileHash(baseDir, srcPath string, watch []string) (hash []byte, matched bool, err error) {
	for _, pattern := range watch {
		relativePath, err := filepath.Rel(baseDir, srcPath)
		if err != nil {
			return nil, false, err
		}
		match, err := doublestar.PathMatch(pattern, relativePath)
		if err != nil {
			return nil, false, err
		}
		if match {
			file, err := os.Open(srcPath)
			if err != nil {
				return nil, false, err
			}

			hasher := sha256.New()
			if _, err := io.Copy(hasher, file); err != nil {
				_ = file.Close()
				return nil, false, err
			}

			hash := hasher.Sum(nil)

			if err := file.Close(); err != nil {
				return nil, false, err
			}
			return hash, true, nil
		}
	}
	return nil, false, nil
}

// computeRootDirs computes the unique root directories for the given baseDir and patterns.
func computeRootDirs(baseDir string, patterns []string) []string {
	uniqueRoots := make(map[string]struct{})
	uniqueRoots[baseDir] = struct{}{}

	for _, pattern := range patterns {
		fullPath := filepath.Join(baseDir, pattern)
		dirPath, _ := doublestar.SplitPattern(fullPath)
		cleanedPath := filepath.Clean(dirPath)

		if _, err := os.Stat(cleanedPath); err == nil {
			uniqueRoots[cleanedPath] = struct{}{}
		}
	}

	roots := make([]string, 0, len(uniqueRoots))
	for root := range uniqueRoots {
		roots = append(roots, root)
	}

	return roots
}
