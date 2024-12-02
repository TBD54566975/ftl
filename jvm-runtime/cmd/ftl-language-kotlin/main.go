package main

import (
	"context"
	"os"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/language/v1/languagepbconnect"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/jvm-runtime/plugin/common"
	"github.com/TBD54566975/ftl/jvm-runtime/plugin/kotlin"
)

func main() {
	plugin.Start(context.Background(),
		os.Getenv("FTL_NAME"),
		createService,
		languagepbconnect.LanguageServiceName,
		languagepbconnect.NewLanguageServiceHandler)
}

func createService(ctx context.Context, config any) (context.Context, *common.Service, error) {
	svc := kotlin.New()
	return ctx, svc, nil
}
