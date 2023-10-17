package internal

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/errors"
)

// UnzipDir unzips a ZIP archive into the specified directory.
func UnzipDir(zipReader *zip.Reader, destDir string) error {
	err := os.MkdirAll(destDir, 0700)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, file := range zipReader.File {
		destPath, err := sanitizeArchivePath(destDir, file.Name)
		if err != nil {
			return errors.WithStack(err)
		}

		// Create directory if it doesn't exist
		if file.FileInfo().IsDir() {
			err := os.MkdirAll(destPath, file.Mode())
			if err != nil {
				return errors.WithStack(err)
			}
			continue
		}

		// Handle symlinks
		if file.Mode()&os.ModeSymlink != 0 {
			reader, err := file.Open()
			if err != nil {
				return errors.WithStack(err)
			}
			buf := &bytes.Buffer{}
			_, err = io.Copy(buf, reader) //nolint:gosec
			if err != nil {
				return errors.WithStack(err)
			}
			// This is probably a little bit aggressive, in that the symlink can
			// only be beneath its parent directory, rather than the root of the
			// zip file. But it's good enough for now.
			symlinkDir := filepath.Dir(destPath)
			symlinkPath, err := sanitizeArchivePath(symlinkDir, buf.String())
			if err != nil {
				return errors.WithStack(err)
			}
			symlinkPath = strings.TrimPrefix(symlinkPath, symlinkDir+"/")
			err = os.Symlink(symlinkPath, destPath)
			if err != nil {
				return errors.WithStack(err)
			}
			continue
		}

		// Handle regular files
		fileReader, err := file.Open()
		if err != nil {
			return errors.WithStack(err)
		}
		defer fileReader.Close()

		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return errors.WithStack(err)
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, fileReader) //nolint:gosec
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func ZipDir(srcDir, destZipFile string) error {
	zipFile, err := os.Create(destZipFile)
	if err != nil {
		return errors.WithStack(err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	return errors.WithStack(filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.WithStack(err)
		}

		// Determine path for the zip file header
		headerPath := strings.TrimPrefix(path, srcDir)
		if strings.HasPrefix(headerPath, string(filepath.Separator)) {
			headerPath = headerPath[1:]
		}

		// Add trailing slash to directory paths
		if info.IsDir() {
			headerPath += "/"
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return errors.WithStack(err)
		}
		header.Name = headerPath

		// Handle symlink
		if info.Mode()&os.ModeSymlink != 0 {
			dest, err := os.Readlink(path)
			if err != nil {
				return errors.WithStack(err)
			}

			header.Method = zip.Store
			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = writer.Write([]byte(dest))
			return errors.WithStack(err)
		}

		// Handle regular files and directories
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return errors.WithStack(err)
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return errors.WithStack(err)
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			return errors.WithStack(err)
		}

		return nil
	}))
}

// Sanitize archive file pathing from "G305: Zip Slip vulnerability"
func sanitizeArchivePath(d, t string) (v string, err error) {
	v = filepath.Join(d, t)
	if strings.HasPrefix(v, filepath.Clean(d)) {
		return v, nil
	}

	return "", fmt.Errorf("%s: %s", "content filepath is tainted", t)
}
