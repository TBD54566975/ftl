package main

import (
	"context"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language/languagepbconnect"
	"github.com/TBD54566975/ftl/common/plugin"
	pythonplugin "github.com/TBD54566975/ftl/python-runtime/python-plugin"
)

func main() {
	plugin.Start(
		context.Background(),
		"ftl-language-python",
		createService,
		languagepbconnect.LanguageServiceName,
		languagepbconnect.NewLanguageServiceHandler,
	)
}

func createService(ctx context.Context, config any) (context.Context, *pythonplugin.Service, error) {
	return ctx, pythonplugin.New(), nil
}
