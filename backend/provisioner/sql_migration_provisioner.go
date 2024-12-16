package provisioner

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/mysql"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver

	"github.com/block/ftl/backend/controller/artefacts"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/sha256"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/dsn"
	"github.com/block/ftl/internal/log"
)

const tenMB = 1024 * 1024 * 10

// NewSQLMigrationProvisioner creates a new provisioner that provisions database migrations
func NewSQLMigrationProvisioner(storage *artefacts.OCIArtefactService) *InMemProvisioner {
	return NewEmbeddedProvisioner(map[schema.ResourceType]InMemResourceProvisionerFn{
		schema.ResourceTypeSQLMigration: provisionSQLMigration(storage),
	})
}

func provisionSQLMigration(storage *artefacts.OCIArtefactService) InMemResourceProvisionerFn {
	return func(ctx context.Context, moduleName string, resource schema.Provisioned) (*RuntimeEvent, error) {
		db, ok := resource.(*schema.Database)
		if !ok {
			return nil, fmt.Errorf("expected database, got %T", resource)
		}
		for migration := range slices.FilterVariants[*schema.MetadataSQLMigration](db.Metadata) {
			parseSHA256, err := sha256.ParseSHA256(migration.Digest)
			if err != nil {
				return nil, fmt.Errorf("failed to parse digest %w", err)
			}
			download, err := storage.Download(ctx, parseSHA256)
			if err != nil {
				return nil, fmt.Errorf("failed to download migration: %w", err)
			}
			dir, err := extractTarToTempDir(download)
			if err != nil {
				return nil, fmt.Errorf("failed to extract tar: %w", err)
			}
			d := ""

			switch db.Type {
			case schema.PostgresDatabaseType:
				d, err = dsn.ResolvePostgresDSN(ctx, db.Runtime.Connections.Write)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve postgres DSN: %w", err)
				}
			case schema.MySQLDatabaseType:
				d, err = dsn.ResolveMySQLDSN(ctx, db.Runtime.Connections.Write)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve mysql DSN: %w", err)
				}
				d = "mysql://" + d
				// strip the tcp part
				exp := regexp.MustCompile(`tcp\((.*?)\)`)
				d = exp.ReplaceAllString(d, "$1")
			}
			u, err := url.Parse(d)
			if err != nil {
				return nil, fmt.Errorf("invalid DSN: %w", err)
			}

			db := dbmate.New(u)
			db.AutoDumpSchema = false
			db.Log = log.FromContext(ctx).Scope("migrate").WriterAt(log.Info)
			db.MigrationsDir = []string{dir}
			err = db.CreateAndMigrate()
			if err != nil {
				return nil, fmt.Errorf("failed to create and migrate database: %w", err)
			}
		}
		return nil, nil
	}
}

func RunMySQLMigration(ctx context.Context, dsn string, moduleDir string, name string) error {
	// strip the tcp part
	exp := regexp.MustCompile(`tcp\((.*?)\)`)
	dsn = exp.ReplaceAllString(dsn, "$1")
	return runDBMateMigration(ctx, "mysql://"+dsn, moduleDir, name)
}

func RunPostgresMigration(ctx context.Context, dsn string, moduleDir string, name string) error {
	return runDBMateMigration(ctx, dsn, moduleDir, name)
}

func runDBMateMigration(ctx context.Context, dsn string, moduleDir string, name string) error {
	migrationDir := filepath.Join(moduleDir, "db", name)
	_, err := os.Stat(migrationDir)
	if err != nil {
		return nil // No migration to run
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return fmt.Errorf("invalid DSN: %w", err)
	}

	db := dbmate.New(u)
	db.AutoDumpSchema = false
	db.Log = log.FromContext(ctx).Scope("migrate").WriterAt(log.Info)
	db.MigrationsDir = []string{migrationDir}
	err = db.CreateAndMigrate()
	if err != nil {
		return fmt.Errorf("failed to create and migrate database: %w", err)
	}
	return nil
}

func extractTarToTempDir(tarReader io.Reader) (tempDir string, err error) {
	// Create a new tar reader
	tr := tar.NewReader(tarReader)

	// Create a temporary directory
	tempDir, err = os.MkdirTemp("", "extracted")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Extract files from the tar archive
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break // End of tar archive
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar header: %w", err)
		}

		// Construct the full path for the file
		targetPath := filepath.Join(tempDir, filepath.Clean(header.Name))

		// Create the file
		file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
		if err != nil {
			return "", fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close()

		// Copy the file content
		if _, err := io.CopyN(file, tr, tenMB); err != nil {
			if !errors.Is(err, io.EOF) {
				return "", fmt.Errorf("failed to copy file content: %w", err)
			}
		}
	}
	return tempDir, nil
}
