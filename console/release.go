//go:build release

package console

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/alecthomas/errors"
)

//go:embed all:dist
var build embed.FS

func Server(ctx context.Context) (http.Handler, error) {
	dir, err := fs.Sub(build, "dist")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = fs.WalkDir(dir, ".", func(path string, d fs.DirEntry, err error) error {
		fmt.Println(path)
		return nil
	})
	return http.FileServer(http.FS(dir)), nil
}
