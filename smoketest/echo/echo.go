// This is the echo module.
package echo

import (
	"context"
	"fmt"
)

// Echo returns a greeting with the current time.
//
//ftl:verb export
func Echo(ctx context.Context, req string) (string, error) {
	return fmt.Sprintf("Hello, %s!!!", req), nil
}
