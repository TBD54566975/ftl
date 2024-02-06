package schema

import (
	"encoding/json"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

var jsonSchemaSample = &Schema{
	Modules: []*Module{
		{Name: "foo", Decls: []Decl{
			&Data{
				Name:     "Foo",
				Comments: []string{"Data comment"},
				Fields: []*Field{
					{Name: "string", Type: &String{}, Comments: []string{"Field comment"}},
					{Name: "int", Type: &Int{}},
					{Name: "float", Type: &Float{}},
					{Name: "optional", Type: &Optional{Type: &String{}}},
					{Name: "bool", Type: &Bool{}},
					{Name: "time", Type: &Time{}},
					{Name: "array", Type: &Array{Element: &String{}}},
					{Name: "arrayOfRefs", Type: &Array{Element: &DataRef{Module: "foo", Name: "Item"}}},
					{Name: "arrayOfArray", Type: &Array{Element: &Array{Element: &String{}}}},
					{Name: "optionalArray", Type: &Array{Element: &Optional{Type: &String{}}}},
					{Name: "map", Type: &Map{Key: &String{}, Value: &Int{}}},
					{Name: "optionalMap", Type: &Map{Key: &String{}, Value: &Optional{Type: &Int{}}}},
					{Name: "ref", Type: &DataRef{Module: "bar", Name: "Bar"}},
					{Name: "any", Type: &Any{}},
					{Name: "keyValue", Type: &DataRef{Module: "foo", Name: "Generic", TypeParameters: []Type{&String{}, &Int{}}}},
				},
			},
			&Data{
				Name: "Item", Fields: []*Field{{Name: "name", Type: &String{}}},
			},
			&Data{
				Name:           "Generic",
				TypeParameters: []*TypeParameter{{Name: "K"}, {Name: "V"}},
				Fields: []*Field{
					{Name: "key", Type: &DataRef{Name: "K"}},
					{Name: "value", Type: &DataRef{Name: "V"}},
				},
			},
		}},
		{Name: "bar", Decls: []Decl{
			&Data{Name: "Bar", Fields: []*Field{{Name: "bar", Type: &String{}}}},
		}},
	},
}

func TestDataToJSONSchema(t *testing.T) {
	schema, err := DataToJSONSchema(jsonSchemaSample, DataRef{Module: "foo", Name: "Foo"})
	assert.NoError(t, err)
	actual, err := json.MarshalIndent(schema, "", "  ")
	assert.NoError(t, err)
	expected := `{
  "description": "Data comment",
  "required": [
    "string",
    "int",
    "float",
    "bool",
    "time",
    "array",
    "arrayOfRefs",
    "arrayOfArray",
    "optionalArray",
    "map",
    "optionalMap",
    "ref",
    "any",
    "keyValue"
  ],
  "additionalProperties": false,
  "definitions": {
    "bar.Bar": {
      "required": [
        "bar"
      ],
      "additionalProperties": false,
      "properties": {
        "bar": {
          "type": "string"
        }
      },
      "type": "object"
    },
    "foo.Generic[String, Int]": {
      "required": [
        "key",
        "value"
      ],
      "additionalProperties": false,
      "properties": {
        "key": {
          "type": "string"
        },
        "value": {
          "type": "integer"
        }
      },
      "type": "object"
    },
    "foo.Item": {
      "required": [
        "name"
      ],
      "additionalProperties": false,
      "properties": {
        "name": {
          "type": "string"
        }
      },
      "type": "object"
    }
  },
  "properties": {
    "any": {},
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
    "arrayOfRefs": {
      "items": {
        "$ref": "#/definitions/foo.Item"
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
    "keyValue": {
      "$ref": "#/definitions/foo.Generic[String, Int]"
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
    "optional": {
      "anyOf": [
        {
          "type": "string"
        },
        {
          "type": "null"
        }
      ]
    },
    "optionalArray": {
      "items": {
        "anyOf": [
          {
            "type": "string"
          },
          {
            "type": "null"
          }
        ]
      },
      "type": "array"
    },
    "optionalMap": {
      "additionalProperties": {
        "anyOf": [
          {
            "type": "integer"
          },
          {
            "type": "null"
          }
        ]
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

func TestJSONSchemaValidation(t *testing.T) {
	input := `
   {
    "string": "string",
    "int": 1,
    "float": 1.23,
    "bool": true,
    "time": "2018-11-13T20:20:39+00:00",
    "array": ["one"],
    "arrayOfRefs": [{"name": "Name"}],
    "arrayOfArray": [[]],
    "optionalArray": [null, "foo"],
    "map": {"one": 2},
    "optionalMap": {"one": 2, "two": null},
    "ref": {"bar": "Name"},
    "any": [{"name": "Name"}, "string", 1, 1.23, true, "2018-11-13T20:20:39+00:00", ["one"], {"one": 2}, null],
    "keyValue": {"key": "string", "value": 1}
  }
   `

	schema, err := DataToJSONSchema(jsonSchemaSample, DataRef{Module: "foo", Name: "Foo"})
	assert.NoError(t, err)
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	assert.NoError(t, err)
	jsonschema, err := jsonschema.CompileString("http://ftl.block.xyz/schema.json", string(schemaJSON))
	assert.NoError(t, err)

	var v interface{}
	err = json.Unmarshal([]byte(input), &v)
	assert.NoError(t, err)

	err = jsonschema.Validate(v)
	assert.NoError(t, err)
}
