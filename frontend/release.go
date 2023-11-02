//go:build release

package frontend

import (
	"context"
	"embed"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/alecthomas/errors"
)

//go:embed all:dist
var build embed.FS

func Server(ctx context.Context, timestamp time.Time, allowOrigin string) (http.Handler, error) {
	dir, err := fs.Sub(build, "dist")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeCORSHeaders(w, allowOrigin)
		var f fs.File
		var err error
		filePath := strings.TrimPrefix(r.URL.Path, "/")
		if ext := path.Ext(filePath); ext != "" {
			f, err = dir.Open(filePath)
		} else {
			// Otherwise return index.html
			f, err = dir.Open("index.html")
		}
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.ServeContent(w, r, filePath, timestamp, f.(io.ReadSeeker))
	}), nil
}
