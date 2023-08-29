package schema

import (
	"fmt"
	"strings"

	"github.com/alecthomas/errors"
	js "github.com/swaggest/jsonschema-go"
)

// DataToJSONSchema converts the schema for a Data object to a JSON Schema.
//
// It takes in the full schema in order to resolve and define references.
func DataToJSONSchema(schema *Schema, dataRef DataRef) (*js.Schema, error) {
	// Collect all data types.
	dataTypes := map[DataRef]*Data{}
	for _, module := range schema.Modules {
		for _, decl := range module.Decls {
			if data, ok := decl.(*Data); ok {
				dataTypes[DataRef{Module: module.Name, Name: data.Name}] = data
			}
		}
	}

	// Find the root data type.
	rootData, ok := dataTypes[dataRef]
	if !ok {
		return nil, errors.Errorf("unknown data type %s", dataRef)
	}

	// Encode root, and collect all data types reachable from the root.
	dataRefs := map[DataRef]bool{}
	root := nodeToJSSchema(rootData, dataRefs)
	if len(dataRefs) == 0 {
		return root, nil
	}
	// Resolve and encode all data types reachable from the root.
	root.Definitions = map[string]js.SchemaOrBool{}
	for dataRef := range dataRefs {
		data, ok := dataTypes[dataRef]
		if !ok {
			return nil, errors.Errorf("unknown data type %s", dataRef)
		}
		root.Definitions[dataRef.String()] = js.SchemaOrBool{TypeObject: nodeToJSSchema(data, dataRefs)}
	}
	return root, nil
}

func nodeToJSSchema(node Node, dataRefs map[DataRef]bool) *js.Schema {
	switch node := node.(type) {
	case *Data:
		st := js.Object
		schema := &js.Schema{
			Description:          jsComments(node.Comments),
			Type:                 &js.Type{SimpleTypes: &st},
			Properties:           map[string]js.SchemaOrBool{},
			AdditionalProperties: jsBool(false),
		}
		for _, field := range node.Fields {
			jsField := nodeToJSSchema(field.Type, dataRefs)
			jsField.Description = jsComments(field.Comments)
			schema.Properties[field.Name] = js.SchemaOrBool{TypeObject: jsField}
		}
		return schema

	case *Int:
		st := js.Integer
		return &js.Schema{Type: &js.Type{SimpleTypes: &st}}

	case *Float:
		st := js.Number
		return &js.Schema{Type: &js.Type{SimpleTypes: &st}}

	case *String:
		st := js.String
		return &js.Schema{Type: &js.Type{SimpleTypes: &st}}

	case *Bool:
		st := js.Boolean
		return &js.Schema{Type: &js.Type{SimpleTypes: &st}}

	case *Time:
		st := js.String
		dt := "date-time"
		return &js.Schema{Type: &js.Type{SimpleTypes: &st}, Format: &dt}

	case *Array:
		st := js.Array
		return &js.Schema{
			Type:  &js.Type{SimpleTypes: &st},
			Items: &js.Items{SchemaOrBool: &js.SchemaOrBool{TypeObject: nodeToJSSchema(node.Element, dataRefs)}},
		}

	case *Map:
		st := js.Object
		// JSON schema generic map of key type to value type
		return &js.Schema{
			Type:                 &js.Type{SimpleTypes: &st},
			AdditionalProperties: &js.SchemaOrBool{TypeObject: nodeToJSSchema(node.Value, dataRefs)},
			PropertyNames:        &js.SchemaOrBool{TypeObject: nodeToJSSchema(node.Key, dataRefs)},
		}

	case *DataRef:
		ref := fmt.Sprintf("#/definitions/%s", node.String())
		dataRefs[*node] = true
		return &js.Schema{Ref: &ref}

	case Decl, *Field, Metadata, *MetadataCalls, *MetadataIngress, *Module, *Schema, Type, *Verb, *VerbRef:
		panic(fmt.Sprintf("unsupported node type %T", node))

	default:
		panic(fmt.Sprintf("unsupported node type %T", node))
	}
}

func jsBool(ok bool) *js.SchemaOrBool {
	return &js.SchemaOrBool{TypeBoolean: &ok}
}

func jsComments(comments []string) *string {
	if len(comments) == 0 {
		return nil
	}
	out := strings.Join(comments, "\n")
	return &out
}
