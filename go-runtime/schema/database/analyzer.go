package database

import (
	"go/ast"
	"go/types"

	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/common/schema"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
)

// Extractor extracts databases to the module schema.
var Extractor = common.NewResourceDeclExtractor[*schema.Database]("database", Extract, matchFunc)

func Extract(pass *analysis.Pass, obj types.Object, node *ast.TypeSpec) optional.Option[*schema.Database] {
	var comments []string
	if md, ok := common.GetFactForObject[*common.ExtractedMetadata](pass, obj).Get(); ok {
		comments = md.Comments
	}
	switch getDBType(pass, node) {
	case postgres:
		return extractDatabase(pass, obj, node, schema.PostgresDatabaseType, comments)
	case mysql:
		return extractDatabase(pass, obj, node, schema.MySQLDatabaseType, comments)
	default:
		return optional.None[*schema.Database]()
	}
}

func extractDatabase(
	pass *analysis.Pass,
	obj types.Object,
	node *ast.TypeSpec,
	dbType string,
	comments []string,
) optional.Option[*schema.Database] {
	db := &schema.Database{
		Pos:      common.GoPosToSchemaPos(pass.Fset, node.Pos()),
		Comments: comments,
		Type:     dbType,
	}

	for _, cfg := range common.GetFactsForObject[*common.DatabaseConfig](pass, obj) {
		if cfg.Method == common.DatabaseConfigMethodName {
			name, ok := cfg.Value.(string)
			if !ok {
				common.Errorf(pass, node, "database name must be a string, was %T", cfg.Value)
				return optional.None[*schema.Database]()
			}
			if !schema.ValidateName(name) {
				common.Errorf(pass, node, "invalid database name %q", name)
				return optional.None[*schema.Database]()
			}
			if name == "" {
				common.Errorf(pass, node.Type, "database config must provide a name")
				return optional.None[*schema.Database]()
			}
			db.Name = name
		}
	}
	// not a DB
	if db.Name == "" {
		return optional.None[*schema.Database]()
	}

	return optional.Some(db)
}

func matchFunc(pass *analysis.Pass, node ast.Node, obj types.Object) bool {
	return getDBType(pass, node) != none
}

type dbType int

const (
	none dbType = iota
	postgres
	mysql
)

func getDBType(pass *analysis.Pass, node ast.Node) dbType {
	ts := node.(*ast.TypeSpec) //nolint:forcetypeassert
	typ, ok := common.GetTypeInfoForNode(ts.Name, pass.TypesInfo).Get()
	if !ok {
		return none
	}
	if common.IsPostgresDatabaseConfigType(pass, typ) {
		return postgres
	}
	if common.IsMysqlDatabaseConfigType(pass, typ) {
		return mysql
	}
	return none
}
