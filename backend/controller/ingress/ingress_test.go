package ingress

import (
	"testing"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/alecthomas/assert/v2"
)

type obj = map[string]any

func TestMatchAndExtractAllSegments(t *testing.T) {
	tests := []struct {
		pattern  string
		urlPath  string
		expected map[string]string
		matched  bool
	}{
		// valid patterns
		{"", "", map[string]string{}, true},
		{"/", "/", map[string]string{}, true},
		{"/{id}", "/123", map[string]string{"id": "123"}, true},
		{"/{id}/{userId}", "/123/456", map[string]string{"id": "123", "userId": "456"}, true},
		{"/users", "/users", map[string]string{}, true},
		{"/users/{id}", "/users/123", map[string]string{"id": "123"}, true},
		{"/users/{id}", "/users/123", map[string]string{"id": "123"}, true},
		{"/users/{id}/posts/{postId}", "/users/123/posts/456", map[string]string{"id": "123", "postId": "456"}, true},

		// invalid patterns
		{"/", "/users", map[string]string{}, false},
		{"/users/{id}", "/bogus/123", map[string]string{}, false},
	}

	for _, test := range tests {
		actual := make(map[string]string)
		match := matchSegments(test.pattern, test.urlPath, func(segment, value string) {
			actual[segment] = value
		})
		assert.Equal(t, test.matched, match, "pattern = %s, urlPath = %s", test.pattern, test.urlPath)
		assert.Equal(t, test.expected, actual, "pattern = %s, urlPath = %s", test.pattern, test.urlPath)
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name    string
		schema  string
		request obj
	}{
		{name: "int", schema: `module test { data Test { intValue Int } }`, request: obj{"intValue": 10.0}},
		{name: "float", schema: `module test { data Test { floatValue Float } }`, request: obj{"floatValue": 10.0}},
		{name: "string", schema: `module test { data Test { stringValue String } }`, request: obj{"stringValue": "test"}},
		{name: "bool", schema: `module test { data Test { boolValue Bool } }`, request: obj{"boolValue": true}},
		{name: "intString", schema: `module test { data Test { intValue Int } }`, request: obj{"intValue": "10"}},
		{name: "floatString", schema: `module test { data Test { floatValue Float } }`, request: obj{"floatValue": "10.0"}},
		{name: "boolString", schema: `module test { data Test { boolValue Bool } }`, request: obj{"boolValue": "true"}},
		{name: "array", schema: `module test { data Test { arrayValue [String] } }`, request: obj{"arrayValue": []any{"test1", "test2"}}},
		{name: "map", schema: `module test { data Test { mapValue {String: String} } }`, request: obj{"mapValue": obj{"key1": "value1", "key2": "value2"}}},
		{name: "dataRef", schema: `module test { data Nested { intValue Int } data Test { dataRef Nested } }`, request: obj{"dataRef": obj{"intValue": 10.0}}},
		{name: "optional", schema: `module test { data Test { intValue Int? } }`, request: obj{}},
		{name: "optionalProvided", schema: `module test { data Test { intValue Int? } }`, request: obj{"intValue": 10.0}},
		{name: "arrayDataRef", schema: `module test { data Nested { intValue Int } data Test { arrayValue [Nested] } }`, request: obj{"arrayValue": []any{obj{"intValue": 10.0}, obj{"intValue": 20.0}}}},
		{name: "mapDataRef", schema: `module test { data Nested { intValue Int } data Test { mapValue {String: Nested} } }`, request: obj{"mapValue": obj{"key1": obj{"intValue": 10.0}, "key2": obj{"intValue": 20.0}}}},
		{name: "otherModuleRef", schema: `module other { data Other { intValue Int } } module test { data Test { otherRef other.Other } }`, request: obj{"otherRef": obj{"intValue": 10.0}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sch, err := schema.ParseString("", test.schema)
			assert.NoError(t, err)

			err = validateRequestMap(&schema.DataRef{Module: "test", Name: "Test"}, nil, test.request, sch)
			assert.NoError(t, err, "%v", test.name)
		})
	}
}
