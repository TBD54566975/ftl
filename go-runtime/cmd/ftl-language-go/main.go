package main

import (
	"context"
	"os"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1/languagepbconnect"
	"github.com/block/ftl/common/plugin"
	"github.com/block/ftl/go-runtime/goplugin"
)

func main() {
	plugin.Start(context.Background(),
		os.Getenv("FTL_NAME"),
		createService,
		languagepbconnect.LanguageServiceName,
		languagepbconnect.NewLanguageServiceHandler)
}

func createService(ctx context.Context, config any) (context.Context, *goplugin.Service, error) {
	svc := goplugin.New()
	return ctx, svc, nil
}
