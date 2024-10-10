// This is the echo module.
package echo

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

var db = ftl.PostgresDatabase("echodb")

// Echo returns a greeting with the current time.
//
//ftl:verb export
func Echo(ctx context.Context, req string) (string, error) {
	_, err := db.Get(ctx).Exec(`CREATE TABLE IF NOT EXISTS messages(
	    message TEXT
	);`)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Hello, %s!!!", req), nil
}
