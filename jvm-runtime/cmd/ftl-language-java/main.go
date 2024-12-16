package main

import (
	"context"
	"os"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1/languagepbconnect"
	"github.com/block/ftl/common/plugin"
	"github.com/block/ftl/jvm-runtime/plugin/common"
	"github.com/block/ftl/jvm-runtime/plugin/java"
)

func main() {
	plugin.Start(context.Background(),
		os.Getenv("FTL_NAME"),
		createService,
		languagepbconnect.LanguageServiceName,
		languagepbconnect.NewLanguageServiceHandler)
}

func createService(ctx context.Context, config any) (context.Context, *common.Service, error) {
	svc := java.New()
	return ctx, svc, nil
}
