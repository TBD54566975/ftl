// This is the echo module.
package echo

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

type EchoDBConfig struct {
	ftl.DefaultPostgresDatabaseConfig
}

func (EchoDBConfig) Name() string { return "echodb" }

// Echo returns a greeting with the current time.
//
//ftl:verb export
func Echo(ctx context.Context, req string) (string, error) {
	return fmt.Sprintf("Hello, %s!!!", req), nil
}
