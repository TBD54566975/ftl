package metadata

import (
	"go/ast"
	"go/token"
	"reflect"

	"github.com/alecthomas/types/optional"
	sets "github.com/deckarep/golang-set/v2"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/TBD54566975/golang-tools/go/analysis/passes/inspect"
	"github.com/TBD54566975/golang-tools/go/ast/inspector"
)

// Extractor extracts metadata to the module schema.
var Extractor = common.NewExtractor("metadata", (*Fact)(nil), Extract)

type Tag struct{} // Tag uniquely identifies the fact type for this extractor.
type Fact = common.DefaultFact[Tag]

func Extract(pass *analysis.Pass) (interface{}, error) {
	in := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector) //nolint:forcetypeassert
	nodeFilter := []ast.Node{
		(*ast.GenDecl)(nil),
		(*ast.FuncDecl)(nil),
	}
	in.Preorder(nodeFilter, func(n ast.Node) {
		var doc *ast.CommentGroup
		switch n := n.(type) {
		case *ast.TypeSpec:
			doc = n.Doc
		case *ast.GenDecl:
			doc = n.Doc
			if doc == nil && len(n.Specs) > 0 {
				if ts, ok := n.Specs[0].(*ast.TypeSpec); ok {
					doc = ts.Doc
				}
			}
		case *ast.FuncDecl:
			doc = n.Doc
		}
		if mdFact, ok := extractMetadata(pass, n, doc).Get(); ok {
			obj, ok := common.GetObjectForNode(pass.TypesInfo, n).Get()
			if !ok {
				return
			}
			common.MarkMetadata(pass, obj, mdFact)
		}
	})
	return common.NewExtractorResult(pass), nil
}

func extractMetadata(pass *analysis.Pass, node ast.Node, doc *ast.CommentGroup) optional.Option[*common.ExtractedMetadata] {
	if doc == nil {
		return optional.None[*common.ExtractedMetadata]()
	}
	directives := common.ParseDirectives(pass, node, doc)
	found := sets.NewSet[string]()
	exported := isExported(directives)
	var declType schema.Decl
	var metadata []schema.Metadata
	for _, dir := range directives {
		var newSchType schema.Decl
		if found.Contains(dir.GetTypeName()) {
			common.Errorf(pass, node, `expected exactly one "ftl:%s" directive but found multiple`,
				dir.GetTypeName())
			continue
		}
		found.Add(dir.GetTypeName())

		if !isAnnotatingValidGoNode(dir, node) {
			if _, ok := node.(*ast.FuncDecl); ok {
				common.NoEndColumnErrorf(pass, dir.GetPosition(), "unexpected directive \"ftl:%s\" attached "+
					"for verb, did you mean to use '//ftl:verb export' instead?", dir.GetTypeName())
				continue
			}

			common.NoEndColumnErrorf(pass, dir.GetPosition(), "unexpected directive \"ftl:%s\"",
				dir.GetTypeName())
			continue
		}

		switch dt := dir.(type) {
		case *common.DirectiveIngress:
			newSchType = &schema.Verb{}
			typ := dt.Type
			if typ == "" {
				typ = "http"
			}
			metadata = append(metadata, &schema.MetadataIngress{
				Pos:    common.GoPosToSchemaPos(pass.Fset, dt.GetPosition()),
				Type:   typ,
				Method: dt.Method,
				Path:   dt.Path,
			})
		case *common.DirectiveCronJob:
			newSchType = &schema.Verb{}
			if exported {
				common.NoEndColumnErrorf(pass, dt.GetPosition(), "ftl:cron cannot be attached to exported verbs")
				continue
			}
			metadata = append(metadata, &schema.MetadataCronJob{
				Pos:  common.GoPosToSchemaPos(pass.Fset, dt.Pos),
				Cron: dt.Cron.String(),
			})
		case *common.DirectiveRetry:
			newSchType = &schema.Verb{}
			metadata = append(metadata, &schema.MetadataRetry{
				Pos:        common.GoPosToSchemaPos(pass.Fset, dt.Pos),
				Count:      dt.Count,
				MinBackoff: dt.MinBackoff,
				MaxBackoff: dt.MaxBackoff,
			})
		case *common.DirectiveSubscriber:
			newSchType = &schema.Verb{}
			metadata = append(metadata, &schema.MetadataSubscriber{
				Pos:  common.GoPosToSchemaPos(pass.Fset, dt.Pos),
				Name: dt.Name,
			})
		case *common.DirectiveVerb:
			newSchType = &schema.Verb{}
		case *common.DirectiveData:
			requireOnlyDirective(pass, node, directives, dt.GetTypeName())
			newSchType = &schema.Data{}
		case *common.DirectiveEnum:
			requireOnlyDirective(pass, node, directives, dt.GetTypeName())
			newSchType = &schema.Enum{}
		case *common.DirectiveTypeAlias:
			requireOnlyDirective(pass, node, directives, dt.GetTypeName())
			newSchType = &schema.TypeAlias{}
		case *common.DirectiveExport:
			requireOnlyDirective(pass, node, directives, dt.GetTypeName())
		}
		declType = updateDeclType(pass, node.Pos(), declType, newSchType)
	}

	return optional.Some(&common.ExtractedMetadata{
		Type:       declType,
		Metadata:   metadata,
		IsExported: exported,
		Comments:   common.ExtractComments(doc),
	})
}

func requireOnlyDirective(pass *analysis.Pass, node ast.Node, directives []common.Directive, typeName string) {
	if len(directives) > 1 {
		common.Errorf(pass, node, "only one directive expected when directive \"ftl:%s\" is present, "+
			"found multiple", typeName)
	}
}

func updateDeclType(pass *analysis.Pass, pos token.Pos, a schema.Decl, b schema.Decl) schema.Decl {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		common.NoEndColumnErrorf(pass, pos, "schema declaration contains conflicting directives")
	}
	return b
}

func isExported(directives []common.Directive) bool {
	for _, d := range directives {
		if exportable, ok := d.(common.Exportable); ok {
			return exportable.IsExported()
		}
	}
	return false
}

func isAnnotatingValidGoNode(dir common.Directive, node ast.Node) bool {
	for _, n := range dir.MustAnnotate() {
		if reflect.TypeOf(n) == reflect.TypeOf(node) {
			return true
		}
	}
	return false
}
