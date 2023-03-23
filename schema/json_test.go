package schema

import (
	"encoding/json"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestJSON(t *testing.T) {
	expected := `{
  "kind": "schema",
  "modules": [
    {
      "kind": "module",
      "comments": [
        "A comment"
      ],
      "decls": [
        {
          "kind": "data",
          "fields": [
            {
              "kind": "field",
              "name": "name",
              "type": {
                "kind": "map",
                "key": {
                  "kind": "string"
                },
                "value": {
                  "kind": "string"
                }
              }
            }
          ],
          "name": "CreateRequest"
        },
        {
          "kind": "data",
          "fields": [
            {
              "kind": "field",
              "name": "name",
              "type": {
                "kind": "array",
                "element": {
                  "kind": "string"
                }
              }
            }
          ],
          "name": "CreateResponse"
        },
        {
          "kind": "data",
          "fields": [
            {
              "kind": "field",
              "comments": [
                "A comment"
              ],
              "name": "name",
              "type": {
                "kind": "string"
              }
            }
          ],
          "name": "DestroyRequest"
        },
        {
          "kind": "data",
          "fields": [
            {
              "kind": "field",
              "name": "name",
              "type": {
                "kind": "string"
              }
            }
          ],
          "name": "DestroyResponse"
        },
        {
          "kind": "verb",
          "metadata": [
            {
              "kind": "metadataCalls",
              "calls": [
                {
                  "kind": "verbRef",
                  "name": "destroy"
                }
              ]
            }
          ],
          "name": "create",
          "request": {
            "kind": "dataRef",
            "name": "CreateRequest"
          },
          "response": {
            "kind": "dataRef",
            "name": "CreateResponse"
          }
        },
        {
          "kind": "verb",
          "name": "destroy",
          "request": {
            "kind": "dataRef",
            "name": "DestroyRequest"
          },
          "response": {
            "kind": "dataRef",
            "name": "DestroyResponse"
          }
        }
      ],
      "name": "todo"
    }
  ]
}`
	data, err := json.MarshalIndent(schema, "", "  ")
	assert.NoError(t, err)
	assert.Equal(t, expected, string(data))
}
