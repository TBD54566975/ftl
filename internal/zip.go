package internal

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// UnzipDir unzips a ZIP archive into the specified directory.
func UnzipDir(zipReader *zip.Reader, destDir string) error {
	err := os.MkdirAll(destDir, 0700)
	if err != nil {
		return err
	}
	for _, file := range zipReader.File {
		destPath := filepath.Clean(filepath.Join(destDir, file.Name)) //nolint:gosec
		if destDir != "." && !strings.HasPrefix(destPath, destDir) {
			return fmt.Errorf("invalid file path: %q", destPath)
		}
		// Create directory if it doesn't exist
		if file.FileInfo().IsDir() {
			err := os.MkdirAll(destPath, file.Mode())
			if err != nil {
				return err
			}
			continue
		}

		// Handle symlinks
		if file.Mode()&os.ModeSymlink != 0 {
			reader, err := file.Open()
			if err != nil {
				return err
			}
			buf := &bytes.Buffer{}
			_, err = io.Copy(buf, reader) //nolint:gosec
			if err != nil {
				return err
			}
			err = os.Symlink(buf.String(), destPath)
			if err != nil {
				return err
			}
			continue
		}

		// Handle regular files
		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, fileReader) //nolint:gosec
		if err != nil {
			return err
		}
	}
	return nil
}

func ZipDir(srcDir, destZipFile string) error {
	zipFile, err := os.Create(destZipFile)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		headerPath := strings.TrimPrefix(path, srcDir)
		if strings.HasPrefix(headerPath, string(filepath.Separator)) {
			headerPath = headerPath[1:]
		}

		if info.IsDir() {
			headerPath += "/"
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = headerPath

		if info.Mode()&os.ModeSymlink != 0 {
			dest, err := os.Readlink(path)
			if err != nil {
				return err
			}

			header.Method = zip.Store
			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return err
			}
			_, err = writer.Write([]byte(dest))
			return err
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			return err
		}

		return nil
	})
}
