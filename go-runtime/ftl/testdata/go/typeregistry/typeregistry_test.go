package typeregistry

import (
	"ftl/builtin"
	"ftl/typeregistry/subpackage"
	"testing"

	"github.com/TBD54566975/ftl/go-runtime/encoding"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/ftl/ftltest"
	"github.com/alecthomas/assert/v2"
)

func TestIngress(t *testing.T) {
	testCases := []struct {
		Name  string
		Input subpackage.StringsTypeEnum
	}{
		{
			Name:  "List",
			Input: subpackage.List([]string{"asdf", "qwerty"}),
		},
		{
			Name:  "Single",
			Input: subpackage.Single("asdf"),
		},
		{
			Name:  "Object",
			Input: subpackage.Object{S: "asdf"},
		},
	}

	ctx := ftltest.Context(ftltest.WithCallsAllowedWithinModule())

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			resp, err := ftl.Call(ctx, Echo, builtin.HttpRequest[EchoRequest]{
				Body: EchoRequest{Strings: test.Input},
			})
			assert.NoError(t, err)
			assert.Equal(t, resp, builtin.HttpResponse[EchoResponse, string]{
				Body: ftl.Some(EchoResponse{Strings: test.Input}),
			})
		})
	}
}

func TestEncoding(t *testing.T) {
	testCases := []struct {
		Name    string
		Input   subpackage.StringsTypeEnum
		Encoded string
	}{
		{
			Name:    "List",
			Input:   subpackage.List([]string{"asdf", "qwerty"}),
			Encoded: `{"input":{"name":"List","value":["asdf","qwerty"]}}`,
		},
		{
			Name:    "Single",
			Input:   subpackage.Single("asdf"),
			Encoded: `{"input":{"name":"Single","value":"asdf"}}`,
		},
		{
			Name:    "Object",
			Input:   subpackage.Object{S: "asdf"},
			Encoded: `{"input":{"name":"Object","value":{"s":"asdf"}}}`,
		},
	}

	type jsonObj struct {
		Input subpackage.StringsTypeEnum
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			input := jsonObj{test.Input}
			jsonBytes, err := encoding.Marshal(input)
			assert.NoError(t, err)
			assert.Equal(t, string(jsonBytes), test.Encoded)
			roundTripOut := jsonObj{}
			err = encoding.Unmarshal(jsonBytes, &roundTripOut)
			assert.NoError(t, err)
			assert.Equal(t, roundTripOut, input)
		})
	}
}
