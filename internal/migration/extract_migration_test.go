package migration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/sha256"
)

func TestExtractMigrations(t *testing.T) {
	t.Run("Valid migrations", func(t *testing.T) {
		// Setup
		migrationsDir := t.TempDir()
		targetDir := t.TempDir()

		// Create dummy migration files
		migrationSubDir := filepath.Join(migrationsDir, "testdb")
		assert.NoError(t, os.Mkdir(migrationSubDir, 0700))
		err := os.WriteFile(filepath.Join(migrationSubDir, "001_init.sql"), []byte("CREATE TABLE test;"), 0600)
		assert.NoError(t, err)

		// Define schema with a database declaration
		db := &schema.Database{Name: "testdb"}
		sch := &schema.Module{Decls: []schema.Decl{db}}

		// Test
		files, err := ExtractSQLMigrations(migrationsDir, targetDir, sch)
		assert.NoError(t, err)

		// Validate results
		targetFile := filepath.Join(targetDir, "testdb.tar")
		assert.Equal(t, targetFile, filepath.Join(targetDir, files[0]))

		// Validate the database metadata
		assert.Equal(t, 1, len(db.Metadata))
		migrationMetadata, ok := db.Metadata[0].(*schema.MetadataSQLMigration)
		assert.True(t, ok)
		expectedDigest, err := sha256.SumFile(targetFile)
		assert.NoError(t, err)
		assert.Equal(t, expectedDigest.String(), migrationMetadata.Digest)
	})

	t.Run("Empty migrations directory", func(t *testing.T) {
		migrationsDir := t.TempDir()
		targetDir := t.TempDir()
		sch := &schema.Module{Decls: []schema.Decl{}}

		files, err := ExtractSQLMigrations(migrationsDir, targetDir, sch)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(files))
	})

	t.Run("Missing migrations directory", func(t *testing.T) {
		migrationsDir := "/non/existent/dir"
		targetDir := t.TempDir()
		sch := &schema.Module{Decls: []schema.Decl{}}

		files, err := ExtractSQLMigrations(migrationsDir, targetDir, sch)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(files))
	})

}
