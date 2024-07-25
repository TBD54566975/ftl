package cors

import (
	"net/http"

	"github.com/rs/cors"
)

func Middleware(allowOrigins []string, allowHeaders []string, next http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins: allowOrigins,
		AllowedHeaders: allowHeaders,
	})
	return c.Handler(next)
}
