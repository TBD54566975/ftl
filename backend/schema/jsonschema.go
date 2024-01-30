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
	data := schema.ResolveDataRef(&dataRef)
	if data == nil {
		return nil, fmt.Errorf("unknown data type %s", dataRef)
	}

	if len(dataRef.TypeParameters) > 0 {
		var err error
		data, err = data.Monomorphise(dataRef.TypeParameters...)
		if err != nil {
			return nil, err
		}
	}

	// Collect all data types.
	dataTypes := schema.DataMap()

	// Encode root, and collect all data types reachable from the root.
	refs := map[Ref]*DataRef{}
	root := nodeToJSSchema(data, refs)
	if len(refs) == 0 {
		return root, nil
	}

	// Resolve and encode all data types reachable from the root.
	root.Definitions = map[string]jsonschema.SchemaOrBool{}
	for key, dataRef := range refs {
		data, ok := dataTypes[Ref{Module: key.Module, Name: key.Name}]
		if !ok {
			return nil, fmt.Errorf("unknown data type %s", key)
		}

		if len(dataRef.TypeParameters) > 0 {
			monomorphisedData, err := data.Monomorphise(dataRef.TypeParameters...)
			if err != nil {
				return nil, err
			}
			data = monomorphisedData

			ref := fmt.Sprintf("%s.%s", dataRef.Module, genericRefName(dataRef))
			root.Definitions[ref] = jsonschema.SchemaOrBool{TypeObject: nodeToJSSchema(data, refs)}
		} else {
			root.Definitions[dataRef.String()] = jsonschema.SchemaOrBool{TypeObject: nodeToJSSchema(data, refs)}
		}

	}
	return root, nil
}

func nodeToJSSchema(node Node, dataRefs map[Ref]*DataRef) *jsonschema.Schema {
	switch node := node.(type) {
	case *Any:
		return &jsonschema.Schema{}

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
		var ref string
		if len(node.TypeParameters) > 0 {
			name := genericRefName(node)
			ref = fmt.Sprintf("#/definitions/%s.%s", node.Module, name)
		} else {
			ref = fmt.Sprintf("#/definitions/%s", node.String())
		}

		dataRefs[node.Untyped()] = node
		schema := &jsonschema.Schema{Ref: &ref}

		return schema

	case *Optional:
		null := jsonschema.Null
		return &jsonschema.Schema{AnyOf: []jsonschema.SchemaOrBool{
			{TypeObject: nodeToJSSchema(node.Type, dataRefs)},
			{TypeObject: &jsonschema.Schema{Type: &jsonschema.Type{SimpleTypes: &null}}},
		}}

	case *TypeParameter:
		return &jsonschema.Schema{}

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

func genericRefName(dataRef *DataRef) string {
	var suffix []string
	for _, t := range dataRef.TypeParameters {
		suffix = append(suffix, t.String())
	}
	return fmt.Sprintf("%s[%s]", dataRef.Name, strings.Join(suffix, ", "))
}
