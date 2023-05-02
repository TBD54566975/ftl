package main

import (
	"context"
	"os"

	"github.com/TBD54566975/ftl/common/plugin"
	drivego "github.com/TBD54566975/ftl/drive-go"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

func main() {
	plugin.Start(context.Background(), os.Getenv("FTL_MODULE"), drivego.Run,
		ftlv1connect.VerbServiceName, ftlv1connect.NewVerbServiceHandler,
		plugin.RegisterAdditionalServer[*drivego.Server](ftlv1connect.DevelServiceName, ftlv1connect.NewDevelServiceHandler))
}
