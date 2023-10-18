package schema

import (
	"fmt"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/swaggest/jsonschema-go"
)

// DataToJSONSchema converts the schema for a Data object to a JSON Schema.
//
// It takes in the full schema in order to resolve and define references.
func DataToJSONSchema(schema *Schema, dataRef DataRef) (*jsonschema.Schema, error) {
	// Collect all data types.
	dataTypes := schema.DataMap()

	// Find the root data type.
	rootData, ok := dataTypes[dataRef]
	if !ok {
		return nil, errors.Errorf("unknown data type %s", dataRef)
	}

	// Encode root, and collect all data types reachable from the root.
	dataRefs := map[DataRef]bool{}
	root := nodeToJSSchema(rootData, dataRef, dataRefs)
	if len(dataRefs) == 0 {
		return root, nil
	}
	// Resolve and encode all data types reachable from the root.
	root.Definitions = map[string]jsonschema.SchemaOrBool{}
	for dataRef := range dataRefs {
		data, ok := dataTypes[dataRef]
		if !ok {
			return nil, errors.Errorf("unknown data type %s", dataRef)
		}
		root.Definitions[dataRef.String()] = jsonschema.SchemaOrBool{TypeObject: nodeToJSSchema(data, dataRef, dataRefs)}
	}
	return root, nil
}

func nodeToJSSchema(node Node, rootRef DataRef, dataRefs map[DataRef]bool) *jsonschema.Schema {
	switch node := node.(type) {
	case *Data:
		st := jsonschema.Object
		schema := &jsonschema.Schema{
			Description:          jsComments(node.Comments),
			Type:                 &jsonschema.Type{SimpleTypes: &st},
			Properties:           map[string]jsonschema.SchemaOrBool{},
			AdditionalProperties: jsBool(false),
		}
		for _, field := range node.Fields {
			jsField := nodeToJSSchema(field.Type, rootRef, dataRefs)
			jsField.Description = jsComments(field.Comments)
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
			Type:  &jsonschema.Type{SimpleTypes: &st},
			Items: &jsonschema.Items{SchemaOrBool: &jsonschema.SchemaOrBool{TypeObject: nodeToJSSchema(node.Element, rootRef, dataRefs)}},
		}

	case *Map:
		st := jsonschema.Object
		// JSON schema generic map of key type to value type
		return &jsonschema.Schema{
			Type:                 &jsonschema.Type{SimpleTypes: &st},
			AdditionalProperties: &jsonschema.SchemaOrBool{TypeObject: nodeToJSSchema(node.Value, rootRef, dataRefs)},
			PropertyNames:        &jsonschema.SchemaOrBool{TypeObject: nodeToJSSchema(node.Key, rootRef, dataRefs)},
		}

	case *DataRef:
		dataRef := *node
		if dataRef.Module == "" {
			// handle root data types
			dataRef.Module = rootRef.Module
		}

		ref := fmt.Sprintf("#/definitions/%s", dataRef.String())
		dataRefs[dataRef] = true
		return &jsonschema.Schema{Ref: &ref}

	case Decl, *Field, Metadata, *MetadataCalls, *MetadataIngress, *Module, *Schema, Type, *Verb, *VerbRef:
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
