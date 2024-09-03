package main

import (
	"fmt"

	"github.com/TBD54566975/ftl/backend/schema"
)

type schemaProtobufCmd struct{}

func (c *schemaProtobufCmd) Run() error { //nolint:unparam
	fmt.Println(schema.ProtobufSchema())
	return nil
}
