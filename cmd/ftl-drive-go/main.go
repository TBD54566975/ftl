package main

import (
	"github.com/TBD54566975/ftl/common/plugin"
	drivego "github.com/TBD54566975/ftl/drive-go"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
)

func main() {
	plugin.Start(drivego.New, ftlv1.RegisterVerbServiceServer,
		plugin.RegisterAdditionalServer[*drivego.Server](ftlv1.RegisterDevelServiceServer))
}
