package main

import (
	"context"
	"os"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1/languagepbconnect"
	"github.com/block/ftl/common/plugin"
	pythonplugin "github.com/block/ftl/python-runtime/python-plugin"
)

func main() {
	plugin.Start(
		context.Background(),
		os.Getenv("FTL_NAME"),
		createService,
		languagepbconnect.LanguageServiceName,
		languagepbconnect.NewLanguageServiceHandler,
	)
}

func createService(ctx context.Context, config any) (context.Context, *pythonplugin.Service, error) {
	return ctx, pythonplugin.New(), nil
}
