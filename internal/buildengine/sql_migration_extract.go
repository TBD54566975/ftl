package buildengine

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/TBD54566975/ftl/common/schema"
	"github.com/TBD54566975/ftl/common/sha256"
	"github.com/TBD54566975/ftl/common/slices"
)

// ExtractSQLMigrations extracts all migrations from the given directory and returns the updated schema and a list of migration files to deploy.
func extractSQLMigrations(migrationsDir string, targetDir string, sch *schema.Module) ([]string, error) {
	paths, err := os.ReadDir(migrationsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}
	ret := []string{}
	for _, dir := range paths {
		path := filepath.Join(migrationsDir, dir.Name())
		dbsDecls := slices.FilterVariants[*schema.Database](sch.Decls)
		for db := range dbsDecls {
			if db.Name != dir.Name() {
				continue
			}
			fileName := db.Name + ".tar"
			target := filepath.Join(targetDir, fileName)
			err := createMigrationTarball(path, target)
			if err != nil {
				return nil, fmt.Errorf("failed to create migration tar %s: %w", dir.Name(), err)
			}
			digest, err := sha256.SumFile(target)
			if err != nil {
				return nil, fmt.Errorf("failed to read migration tar for sha256 %s: %w", dir.Name(), err)
			}
			db.Metadata = append(db.Metadata, &schema.MetadataSQLMigration{Digest: digest.String()})
			if err != nil {
				return nil, fmt.Errorf("failed to read migrations for %s: %w", dir.Name(), err)
			}
			ret = append(ret, fileName)
			break
		}
	}
	return ret, nil
}

func createMigrationTarball(migrationDir string, target string) error {
	// Create the tar file
	tarFile, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("failed to create tar file: %w", err)
	}
	defer tarFile.Close()

	// Create a new tar writer
	tw := tar.NewWriter(tarFile)
	defer tw.Close()

	// Read the directory
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Sort files alphabetically
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	// Set the Unix epoch time
	epoch := time.Unix(0, 0)

	// Add files to the tarball
	for _, file := range files {

		filePath := filepath.Join(migrationDir, file.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("failed to stat file: %w", err)
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return fmt.Errorf("failed to create tar header: %w", err)
		}
		header.Name = file.Name()
		header.ModTime = epoch
		header.AccessTime = epoch
		header.ChangeTime = epoch

		// Write header
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}

		// Write file content
		if !info.IsDir() {
			fileContent, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer fileContent.Close()

			if _, err := io.Copy(tw, fileContent); err != nil {
				return fmt.Errorf("failed to write file content: %w", err)
			}
		}
	}
	return nil
}

func handleDatabaseMigrations(deployDir string, dbDir string, module *schema.Module) ([]string, error) {
	target := filepath.Join(deployDir, ".ftl", "migrations")
	err := os.MkdirAll(target, 0770) // #nosec
	if err != nil {
		return nil, fmt.Errorf("failed to create migration directory: %w", err)
	}
	migrations, err := extractSQLMigrations(dbDir, target, module)
	if err != nil {
		return nil, fmt.Errorf("failed to extract migrations: %w", err)
	}
	relativeFiles := []string{}
	for _, file := range migrations {
		filePath := filepath.Join(".ftl", "migrations", file)
		relativeFiles = append(relativeFiles, filePath)
	}
	return relativeFiles, nil
}
