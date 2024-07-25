package cron

import (
	"context"
	"os"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type EchoRequest struct {
	Name ftl.Option[string] `json:"name"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

//ftl:cron * * * * * * *
func Cron(ctx context.Context) error {
	return os.WriteFile(os.Getenv("DEST_FILE"), []byte("Hello, world!"), 0644)
}

//ftl:cron 5m
func FiveMinutes(ctx context.Context) error {
	return nil
}

//ftl:cron Sat
func Saturday(ctx context.Context) error {
	return nil
}
