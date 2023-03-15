package main

import (
	ftlv1 "github.com/TBD54566975/ftl/common/gen/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/common/plugin"
	drivego "github.com/TBD54566975/ftl/drive-go"
)

func main() {
	plugin.Start(drivego.New, ftlv1.RegisterVerbServiceServer,
		plugin.RegisterAdditionalServer[*drivego.Server](ftlv1.RegisterDevelServiceServer))
}
