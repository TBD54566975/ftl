package schema

import (
	"fmt"
	"strings"

	"github.com/swaggest/jsonschema-go"
)

// DataToJSONSchema converts the schema for a Data object to a JSON Schema.
//
// It takes in the full schema in order to resolve and define references.
func DataToJSONSchema(schema *Schema, dataRef DataRef) (*jsonschema.Schema, error) {
	// Collect all data types.
	dataTypes := schema.DataMap()

	// Find the root data type.
	rootData, ok := dataTypes[DataRef{Module: dataRef.Module, Name: dataRef.Name}]
	if !ok {
		return nil, fmt.Errorf("unknown data type %s", dataRef)
	}

	// Encode root, and collect all data types reachable from the root.
	dataRefs := map[DataRef]bool{}
	root := nodeToJSSchema(rootData, dataRefs)
	if len(dataRefs) == 0 {
		return root, nil
	}
	// Resolve and encode all data types reachable from the root.
	root.Definitions = map[string]jsonschema.SchemaOrBool{}
	for dataRef := range dataRefs {
		data, ok := dataTypes[DataRef{Module: dataRef.Module, Name: dataRef.Name}]
		if !ok {
			return nil, fmt.Errorf("unknown data type %s", dataRef)
		}
		root.Definitions[dataRef.String()] = jsonschema.SchemaOrBool{TypeObject: nodeToJSSchema(data, dataRefs)}
	}
	return root, nil
}

func nodeToJSSchema(node Node, dataRefs map[DataRef]bool) *jsonschema.Schema {
	switch node := node.(type) {
	case *Unit:
		st := jsonschema.Object
		return &jsonschema.Schema{Type: &jsonschema.Type{SimpleTypes: &st}}

	case *Data:
		st := jsonschema.Object
		schema := &jsonschema.Schema{
			Description:          jsComments(node.Comments),
			Type:                 &jsonschema.Type{SimpleTypes: &st},
			Properties:           map[string]jsonschema.SchemaOrBool{},
			AdditionalProperties: jsBool(false),
		}
		for _, field := range node.Fields {
			jsField := nodeToJSSchema(field.Type, dataRefs)
			jsField.Description = jsComments(field.Comments)
			if _, ok := field.Type.(*Optional); !ok {
				schema.Required = append(schema.Required, field.Name)
			}
			schema.Properties[field.Name] = jsonschema.SchemaOrBool{TypeObject: jsField}
		}
		return schema

	case *Int:
		st := jsonschema.Integer
		return &jsonschema.Schema{Type: &jsonschema.Type{SimpleTypes: &st}}

	case *Float:
		st := jsonschema.Number
		return &jsonschema.Schema{Type: &jsonschema.Type{SimpleTypes: &st}}

	case *String:
		st := jsonschema.String
		return &jsonschema.Schema{Type: &jsonschema.Type{SimpleTypes: &st}}

	case *Bytes:
		st := jsonschema.String
		encoding := "base64"
		mediaType := "application/octet-stream"
		return &jsonschema.Schema{
			Type:             &jsonschema.Type{SimpleTypes: &st},
			ContentEncoding:  &encoding,
			ContentMediaType: &mediaType,
		}

	case *Bool:
		st := jsonschema.Boolean
		return &jsonschema.Schema{Type: &jsonschema.Type{SimpleTypes: &st}}

	case *Time:
		st := jsonschema.String
		dt := "date-time"
		return &jsonschema.Schema{Type: &jsonschema.Type{SimpleTypes: &st}, Format: &dt}

	case *Array:
		st := jsonschema.Array
		return &jsonschema.Schema{
			Type: &jsonschema.Type{SimpleTypes: &st},
			Items: &jsonschema.Items{
				SchemaOrBool: &jsonschema.SchemaOrBool{
					TypeObject: nodeToJSSchema(node.Element, dataRefs),
				},
			},
		}

	case *Map:
		st := jsonschema.Object
		// JSON schema generic map of key type to value type
		return &jsonschema.Schema{
			Type:                 &jsonschema.Type{SimpleTypes: &st},
			PropertyNames:        &jsonschema.SchemaOrBool{TypeObject: nodeToJSSchema(node.Key, dataRefs)},
			AdditionalProperties: &jsonschema.SchemaOrBool{TypeObject: nodeToJSSchema(node.Value, dataRefs)},
		}

	case *DataRef:
		dataRef := *node
		ref := fmt.Sprintf("#/definitions/%s", dataRef.String())
		dataRefs[dataRef] = true
		return &jsonschema.Schema{Ref: &ref}

	case *Optional:
		null := jsonschema.Null
		return &jsonschema.Schema{AnyOf: []jsonschema.SchemaOrBool{
			{TypeObject: nodeToJSSchema(node.Type, dataRefs)},
			{TypeObject: &jsonschema.Schema{Type: &jsonschema.Type{SimpleTypes: &null}}},
		}}

	case Decl, *Field, Metadata, *MetadataCalls, *MetadataDatabases, *MetadataIngress,
		IngressPathComponent, *IngressPathLiteral, *IngressPathParameter, *Module,
		*Schema, Type, *Database, *Verb, *VerbRef, *SourceRef, *SinkRef:
		panic(fmt.Sprintf("unsupported node type %T", node))

	default:
		panic(fmt.Sprintf("unsupported node type %T", node))
	}
}

func jsBool(ok bool) *jsonschema.SchemaOrBool {
	return &jsonschema.SchemaOrBool{TypeBoolean: &ok}
}

func jsComments(comments []string) *string {
	if len(comments) == 0 {
		return nil
	}
	out := strings.Join(comments, "\n")
	return &out
}
