package schema

import (
	"fmt"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/schema/analyzers"
	"github.com/TBD54566975/golang-tools/go/analysis"
	checker "github.com/TBD54566975/golang-tools/go/analysis/programmaticchecker"
	"github.com/TBD54566975/golang-tools/go/packages"
)

// Extract statically parses Go FTL module source into a schema.Module
func Extract(moduleDir string) (analyzers.ExtractResult, error) {
	pkgConfig := packages.Config{
		Dir:  moduleDir,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
	}
	cConfig := checker.Config{
		LoadConfig:           pkgConfig,
		RunDespiteLoadErrors: true,
		Patterns:             []string{"./..."},
	}
	results, diagnostics, err := checker.Run(cConfig, append(analyzers.Extractors, analyzers.Finalizer)...)
	if err != nil {
		return analyzers.ExtractResult{}, err
	}
	fResult, ok := results[analyzers.Finalizer]
	if !ok {
		return analyzers.ExtractResult{}, fmt.Errorf("schema extraction finalizer result not found")
	}

	if len(fResult) == 0 {
		return analyzers.ExtractResult{}, fmt.Errorf("schema extraction finalizer result is empty")
	}

	r, ok := fResult[0].(analyzers.ExtractResult)
	if !ok {
		return analyzers.ExtractResult{}, fmt.Errorf("unexpected schema extraction result type: %T", fResult[0])
	}

	errors := diagnosticsToSchemaErrors(diagnostics)
	schema.SortErrorsByPosition(errors)
	r.Errors = errors

	return r, nil
}

func diagnosticsToSchemaErrors(diagnostics []analysis.SimpleDiagnostic) []*schema.Error {
	if len(diagnostics) == 0 {
		return nil
	}
	errors := make([]*schema.Error, 0, len(diagnostics))
	for _, d := range diagnostics {
		errors = append(errors, &schema.Error{
			Pos:       simplePosToSchemaPos(d.Pos),
			EndColumn: d.End.Column,
			Msg:       d.Message,
			Level:     analyzers.DiagnosticCategory(d.Category).ToErrorLevel(),
		})
	}
	return errors
}

func simplePosToSchemaPos(pos analysis.SimplePosition) schema.Position {
	return schema.Position{
		Filename: pos.Filename,
		Offset:   pos.Offset,
		Line:     pos.Line,
		Column:   pos.Column,
	}
}
