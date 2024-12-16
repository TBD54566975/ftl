//go:build release

package console

import (
	"context"
	"embed"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"errors"

	"github.com/block/ftl/internal/cors"
)

//go:embed all:dist
var build embed.FS

func Server(ctx context.Context, timestamp time.Time, allowOrigin *url.URL) (http.Handler, error) {
	dir, err := fs.Sub(build, "dist")
	if err != nil {
		return nil, err
	}
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})
	if allowOrigin != nil {
		handler = cors.Middleware([]string{allowOrigin.String()}, nil, handler)
	}
	return handler, nil
}
