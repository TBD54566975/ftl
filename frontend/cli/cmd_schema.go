package main

type schemaCmd struct {
	Get      getSchemaCmd      `default:"" cmd:"" help:"Retrieve the cluster FTL schema."`
	Diff     schemaDiffCmd     `cmd:"" help:"Print any schema differences between this cluster and another cluster. Returns an exit code of 1 if there are differences."`
	Generate schemaGenerateCmd `cmd:"" help:"Stream the schema from the cluster and generate files from the template."`
	Import   schemaImportCmd   `cmd:"" help:"Import messages to the FTL schema."`
}
