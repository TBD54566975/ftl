package dsn

import "fmt"

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
