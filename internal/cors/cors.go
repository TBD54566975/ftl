package cors

import (
	"net/http"

	"github.com/rs/cors"
)

func Middleware(allowOrigins []string, next http.Handler) http.Handler {
	c := cors.New(cors.Options{AllowedOrigins: allowOrigins})
	return c.Handler(next)
}
