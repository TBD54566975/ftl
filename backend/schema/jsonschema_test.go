package schema

import (
	"encoding/json"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestDataToJSONSchema(t *testing.T) {
	schema, err := DataToJSONSchema(&Schema{
		Modules: []*Module{
			{Name: "foo", Decls: []Decl{&Data{
				Name:     "Foo",
				Comments: []string{"Data comment"},
				Fields: []*Field{
					{Name: "string", Type: &String{}, Comments: []string{"Field comment"}},
					{Name: "int", Type: &Int{}},
					{Name: "float", Type: &Float{}},
					{Name: "bool", Type: &Bool{}},
					{Name: "time", Type: &Time{}},
					{Name: "array", Type: &Array{Element: &String{}}},
					{Name: "arrayOfArray", Type: &Array{Element: &Array{Element: &String{}}}},
					{Name: "map", Type: &Map{Key: &String{}, Value: &Int{}}},
					{Name: "ref", Type: &DataRef{Module: "bar", Name: "Bar"}},
				}}}},
			{Name: "bar", Decls: []Decl{
				&Data{Name: "Bar", Fields: []*Field{{Name: "bar", Type: &String{}}}},
			}},
		},
	}, DataRef{Module: "foo", Name: "Foo"})
	assert.NoError(t, err)
	actual, err := json.MarshalIndent(schema, "", "  ")
	assert.NoError(t, err)
	expected := `{
  "description": "Data comment",
  "additionalProperties": false,
  "definitions": {
    "bar.Bar": {
      "additionalProperties": false,
      "properties": {
        "bar": {
          "type": "string"
        }
      },
      "type": "object"
    }
  },
  "properties": {
    "array": {
      "items": {
        "type": "string"
      },
      "type": "array"
    },
    "arrayOfArray": {
      "items": {
        "items": {
          "type": "string"
        },
        "type": "array"
      },
      "type": "array"
    },
    "bool": {
      "type": "boolean"
    },
    "float": {
      "type": "number"
    },
    "int": {
      "type": "integer"
    },
    "map": {
      "additionalProperties": {
        "type": "integer"
      },
      "propertyNames": {
        "type": "string"
      },
      "type": "object"
    },
    "ref": {
      "$ref": "#/definitions/bar.Bar"
    },
    "string": {
      "description": "Field comment",
      "type": "string"
    },
    "time": {
      "type": "string",
      "format": "date-time"
    }
  },
  "type": "object"
}`
	assert.Equal(t, expected, string(actual))
}
