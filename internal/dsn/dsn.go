package dsn

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"

	"github.com/block/ftl/common/schema"
)

type dsnOptions struct {
	host string
	port int
}

type Option func(*dsnOptions)

func Port(port int) Option {
	return func(o *dsnOptions) {
		o.port = port
	}
}

func Host(host string) Option {
	return func(o *dsnOptions) {
		o.host = host
	}
}

// PostgresDSN returns a PostgresDSN string for connecting to the FTL Controller PG database.
func PostgresDSN(dbName string, options ...Option) string {
	opts := &dsnOptions{port: 15432, host: "127.0.0.1"}
	for _, opt := range options {
		opt(opts)
	}
	return fmt.Sprintf("postgres://%s:%d/%s?sslmode=disable&user=postgres&password=secret", opts.host, opts.port, dbName)
}

// MySQLDSN returns a MySQLDSN string for connecting to the local MySQL database.
func MySQLDSN(dbName string, options ...Option) string {
	opts := &dsnOptions{port: 13306, host: "127.0.0.1"}
	for _, opt := range options {
		opt(opts)
	}
	return fmt.Sprintf("root:secret@tcp(%s:%d)/%s?allowNativePasswords=True", opts.host, opts.port, dbName)
}

func ResolvePostgresDSN(ctx context.Context, connector schema.DatabaseConnector) (string, error) {
	switch c := connector.(type) {
	case *schema.DSNDatabaseConnector:
		return c.DSN, nil
	case *schema.AWSIAMAuthDatabaseConnector:
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return "", fmt.Errorf("configuration error: %w", err)
		}

		region, err := parseRegionFromEndpoint(c.Endpoint)
		if err != nil {
			return "", fmt.Errorf("failed to parse region from endpoint: %w", err)
		}

		authenticationToken, err := auth.BuildAuthToken(ctx, c.Endpoint, region, c.Username, cfg.Credentials)
		if err != nil {
			return "", fmt.Errorf("failed to create authentication token: %w", err)
		}
		host, port, err := net.SplitHostPort(c.Endpoint)
		if err != nil {
			return "", fmt.Errorf("failed to split host and port: %w", err)
		}
		return fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s", host, port, c.Database, c.Username, authenticationToken), nil
	default:
		return "", fmt.Errorf("unexpected database connector type: %T", connector)
	}
}

func parseRegionFromEndpoint(endpoint string) (string, error) {
	host, _, err := net.SplitHostPort(endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to split host and port: %w", err)
	}
	host = strings.TrimSuffix(host, ".rds.amazonaws.com")
	parts := strings.Split(host, ".")
	return parts[len(parts)-1], nil
}

func ResolveMySQLDSN(ctx context.Context, connector schema.DatabaseConnector) (string, error) {
	dsnRuntime, ok := connector.(*schema.DSNDatabaseConnector)
	if !ok {
		return "", fmt.Errorf("unexpected database connector type: %T", connector)
	}
	return dsnRuntime.DSN, nil
}
