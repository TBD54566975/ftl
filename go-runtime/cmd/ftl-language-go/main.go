package main

import (
	"context"
	"os"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language/languagepbconnect"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/go-runtime/goplugin"
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
