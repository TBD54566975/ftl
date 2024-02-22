package buildengine

import (
	"bufio"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/TBD54566975/ftl/internal"
)

// ErrSkip can be returned by the WalkDir callback to skip a file or directory.
var ErrSkip = errors.New("skip directory")

// WalkDir performs a depth-first walk of the file tree rooted at dir, calling
// fn for each file or directory in the tree, including dir.
//
// It will adhere to .gitignore files. The callback "fn" can return ErrSkip to
// skip recursion.
func WalkDir(dir string, fn func(path string, d fs.DirEntry) error) error {
	return walkDir(dir, initGitIgnore(dir), fn)
}

// Depth-first walk of dir executing fn after each entry.
func walkDir(dir string, ignores []string, fn func(path string, d fs.DirEntry) error) error {
	dirInfo, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if err = fn(dir, fs.FileInfoToDirEntry(dirInfo)); err != nil {
		if errors.Is(err, ErrSkip) {
			return nil
		}
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var dirs []os.DirEntry

	// Process files first, then recurse into directories.
	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())

		// Check if the path matches any ignore pattern
		shouldIgnore := false
		for _, pattern := range ignores {
			match, err := doublestar.PathMatch(pattern, fullPath)
			if err != nil {
				return err
			}
			if match {
				shouldIgnore = true
				break
			}
		}

		if shouldIgnore {
			continue // Skip this entry
		}

		if entry.IsDir() {
			dirs = append(dirs, entry)
		} else {
			if err = fn(fullPath, entry); err != nil {
				if errors.Is(err, ErrSkip) {
					// If errSkip is found in a file, skip the remaining files in this directory
					return nil
				}
				return err
			}
		}
	}

	// Then, recurse into subdirectories
	for _, dirEntry := range dirs {
		dirPath := filepath.Join(dir, dirEntry.Name())
		ignores = append(ignores, loadGitIgnore(dirPath)...)
		if err := walkDir(dirPath, ignores, fn); err != nil {
			if errors.Is(err, ErrSkip) {
				return ErrSkip // Propagate errSkip upwards to stop this branch of recursion
			}
			return err
		}
	}
	return nil
}

func initGitIgnore(dir string) []string {
	ignore := []string{
		"**/.*",
		"**/.*/**",
	}
	home, err := os.UserHomeDir()
	if err == nil {
		ignore = append(ignore, loadGitIgnore(home)...)
	}
	gitRoot := internal.GitRoot(dir)
	if gitRoot != "" {
		for current := dir; strings.HasPrefix(current, gitRoot); current = path.Dir(current) {
			ignore = append(ignore, loadGitIgnore(current)...)
		}
	}
	return ignore
}

func loadGitIgnore(dir string) []string {
	r, err := os.Open(path.Join(dir, ".gitignore"))
	if err != nil {
		return nil
	}
	ignore := []string{}
	lr := bufio.NewScanner(r)
	for lr.Scan() {
		line := lr.Text()
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' || line[0] == '!' { // We don't support negation.
			continue
		}
		if strings.HasSuffix(line, "/") {
			line = path.Join("**", line, "**/*")
		} else if !strings.ContainsRune(line, '/') {
			line = path.Join("**", line)
		}
		ignore = append(ignore, line)
	}
	return ignore
}
