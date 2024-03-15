package schema

import (
	"fmt"
	"strings"

	"github.com/swaggest/jsonschema-go"
)

// DataToJSONSchema converts the schema for a Data object to a JSON Schema.
//
// It takes in the full schema in order to resolve and define references.
func DataToJSONSchema(sch *Schema, ref Ref) (*jsonschema.Schema, error) {
	data, err := sch.ResolveRefMonomorphised(&ref)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("unknown data type %s", ref)
	}

	// Collect all data types.
	dataTypes := sch.DataMap()

	// Encode root, and collect all data types reachable from the root.
	refs := map[RefKey]*Ref{}
	root := nodeToJSSchema(sch, data, refs)
	if len(refs) == 0 {
		return root, nil
	}

	// Resolve and encode all data types reachable from the root.
	root.Definitions = map[string]jsonschema.SchemaOrBool{}
	for key, r := range refs {
		data, ok := dataTypes[RefKey{Module: key.Module, Name: key.Name}]
		if !ok {
			return nil, fmt.Errorf("unknown data type %s", key)
		}

		if len(r.TypeParameters) > 0 {
			monomorphisedData, err := data.Monomorphise(r)
			if err != nil {
				return nil, err
			}
			data = monomorphisedData

			ref := fmt.Sprintf("%s.%s", r.Module, refName(r))
			root.Definitions[ref] = jsonschema.SchemaOrBool{TypeObject: nodeToJSSchema(sch, data, refs)}
		} else {
			root.Definitions[r.String()] = jsonschema.SchemaOrBool{TypeObject: nodeToJSSchema(sch, data, refs)}
		}

	}
	return root, nil
}

func nodeToJSSchema(sch *Schema, node Node, dataRefs map[RefKey]*Ref) *jsonschema.Schema {
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
			jsField := nodeToJSSchema(sch, field.Type, dataRefs)
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
					TypeObject: nodeToJSSchema(sch, node.Element, dataRefs),
				},
			},
		}

	case *Map:
		st := jsonschema.Object
		// JSON schema generic map of key type to value type
		return &jsonschema.Schema{
			Type:                 &jsonschema.Type{SimpleTypes: &st},
			PropertyNames:        &jsonschema.SchemaOrBool{TypeObject: nodeToJSSchema(sch, node.Key, dataRefs)},
			AdditionalProperties: &jsonschema.SchemaOrBool{TypeObject: nodeToJSSchema(sch, node.Value, dataRefs)},
		}

	case *Ref:
		var ref string
		if len(node.TypeParameters) > 0 {
			name := refName(node)
			ref = fmt.Sprintf("#/definitions/%s.%s", node.Module, name)
		} else {
			ref = fmt.Sprintf("#/definitions/%s", node.String())
		}

		decl := sch.ResolveRef(node)
		if decl != nil {
			if _, ok := decl.(*Data); ok {
				dataRefs[node.ToRefKey()] = node
			}
		}

		schema := &jsonschema.Schema{Ref: &ref}

		return schema

	case *Optional:
		null := jsonschema.Null
		return &jsonschema.Schema{AnyOf: []jsonschema.SchemaOrBool{
			{TypeObject: nodeToJSSchema(sch, node.Type, dataRefs)},
			{TypeObject: &jsonschema.Schema{Type: &jsonschema.Type{SimpleTypes: &null}}},
		}}

	case *TypeParameter:
		return &jsonschema.Schema{}

	case Decl, *Field, Metadata, *MetadataCalls, *MetadataDatabases, *MetadataIngress,
		*MetadataAlias, IngressPathComponent, *IngressPathLiteral, *IngressPathParameter, *Module,
		*Schema, Type, *Database, *Verb, *Enum, *EnumVariant,
		Value, *StringValue, *IntValue, *Config, *Secret, Symbol:
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

func refName(ref *Ref) string {
	var suffix []string
	for _, t := range ref.TypeParameters {
		suffix = append(suffix, t.String())
	}
	return fmt.Sprintf("%s[%s]", ref.Name, strings.Join(suffix, ", "))
}
