package main

type schemaCmd struct {
	Get      getSchemaCmd      `default:"" cmd:"" help:"Retrieve the cluster FTL schema."`
	Protobuf schemaProtobufCmd `cmd:"" help:"Generate protobuf schema mirroring the FTL schema structure."`
	Generate schemaGenerateCmd `cmd:"" help:"Stream the schema from the cluster and generate files from the template."`
}
