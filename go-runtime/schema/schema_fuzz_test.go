package schema

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/alecthomas/assert/v2"
)

var symbols = []string{
	"any", "bool", "[]byte", "float64", "int", "string", "time.Time", "ftl.Unit",
}

// generate all possible permutations for slices, maps and ftl.Option[...]
func generatePermutations(symbols []string) []string {
	permutations := append([]string{}, symbols...)
	for _, symbol := range symbols {
		permutations = append(permutations, "[]"+symbol)
		permutations = append(permutations, "ftl.Option["+symbol+"]")
		// don't add slices as map keys
		if strings.HasPrefix(symbol, "[]") {
			continue
		}
		for _, nestedSymbol := range symbols {
			permutations = append(permutations, fmt.Sprintf("map[%s]%s", symbol, nestedSymbol))
		}
	}
	return permutations
}

func generateSymbolTypeStrings(symbols []string) map[string]string {
	symbolTypeStringMap := make(map[string]string)
	for _, symbol := range symbols {
		symbolTypeStringMap[symbol] = getSchemaType(symbol).String()
	}
	return symbolTypeStringMap
}

func getSchemaType(symbol string) schema.Type {
	switch symbol {
	case "any":
		return &schema.Any{}
	case "bool":
		return &schema.Bool{}
	case "[]byte":
		return &schema.Bytes{}
	case "float64":
		return &schema.Float{}
	case "int":
		return &schema.Int{}
	case "string":
		return &schema.String{}
	case "time.Time":
		return &schema.Time{}
	case "ftl.Unit":
		return &schema.Unit{}
	default:
		if strings.HasPrefix(symbol, "[]") {
			// `[]` is 2 characters long
			return &schema.Array{Element: getSchemaType(symbol[2:])}
		}
		if strings.HasPrefix(symbol, "map[") {
			key := symbol[4:findMatchingBracketIndex(symbol, 3)]
			value := symbol[4+len(key)+1:] // remainder of the string after the key
			return &schema.Map{Key: getSchemaType(key), Value: getSchemaType(value)}
		}
		if strings.HasPrefix(symbol, "ftl.Option[") {
			// `ftl.Option[` is 11 characters long
			return &schema.Optional{Type: getSchemaType(symbol[11 : len(symbol)-1])}
		}
		panic(fmt.Sprintf("unexpected symbol: %s", symbol))
	}
}

func findMatchingBracketIndex(symbol string, startIdx int) int {
	bracketCount := 1
	for i := startIdx + 1; i < len(symbol); i++ {
		switch symbol[i] {
		case '[':
			bracketCount++
		case ']':
			bracketCount--
			if bracketCount == 0 {
				return i
			}
		}
	}
	return -1
}

func generateSourceCode(symbol string) string {
	return `package test

import (
	"context"
	` + (func() string {
		if strings.Contains(symbol, "time.Time") {
			return "\t\"time\"\n"
		}
		return ""
	}()) + `
	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

var config = ftl.Config[` + symbol + `]("cfg")

var secret = ftl.Secret[` + symbol + `]("secret")

//ftl:data
type Data struct {
    DataField ` + symbol + `
    ParamDataField ParameterizedData[` + symbol + `]
}

type ParameterizedData[T any] struct {
    Field T
}

//ftl:data export
type ExportedData struct {
    Field string
}

//ftl:verb
func DataFunc(ctx context.Context, req Data) (Data, error) {
    return Data{}, nil
}


var db = ftl.PostgresDatabase("testDb")

` + (func() string {
		if symbol == "int" || symbol == "string" {
			return `

//ftl:enum
type Color ` + symbol + `
const (
	` + (func() string {
				switch symbol {
				case "int":
					return `Red   Color = iota
	Blue
	Green`
				case "string":
					return `Red   Color = "Red"
	Blue  Color = "Blue"
	Green Color = "Green"`
				}
				return ""
			}()) + `
)
`
		}
		return ""
	}()) + `

` + (func() string {
		if symbol != "any" {
			return `//ftl:enum
type Discriminator interface {
    tag()
}

type Variant ` + symbol + `
func (Variant) tag() {}

//ftl:enum export
type ExportedDiscriminator interface {
    exportedTag()
}

type ExportedVariant ` + symbol + `
func (ExportedVariant) exportedTag() {}
`
		}
		return ""
	}()) + `

var Topic = ftl.Topic[` + symbol + `]("topic")

//ftl:export
var ExportedTopic = ftl.Topic[` + symbol + `]("exported_topic")

var _ = ftl.Subscription(Topic, "subscription")

//ftl:typealias
type Alias ` + symbol + `

//ftl:typealias export
type ExportedAlias ` + symbol + `

//ftl:typealias
type EqualAlias = ` + symbol + `

//ftl:typealias export
type ExportedEqualAlias = ` + symbol + `

//ftl:verb
func Func(ctx context.Context, req ` + symbol + `) (` + symbol + `, error) {
    panic("not implemented")
}

//ftl:verb export
func ExportedFunc(ctx context.Context, req ` + symbol + `) (` + symbol + `, error) {
    panic("not implemented")
}

//ftl:verb
func SourceFunc(ctx context.Context) (` + symbol + `, error) {
    panic("not implemented")
}

//ftl:verb
func SinkFunc(ctx context.Context, req ` + symbol + `) error {
    panic("not implemented")
}

//ftl:verb
func EmptyVerbFunc(ctx context.Context) error {
    return nil
}
`
}

func FuzzExtract(f *testing.F) {
	if testing.Short() {
		f.Skip("skipping test in short mode")
	}

	allSymbols := generatePermutations(symbols)
	for _, symbol := range allSymbols {
		f.Add(symbol)
	}
	typenames := generateSymbolTypeStrings(allSymbols)

	f.Fuzz(func(t *testing.T, symbolType string) {
		code := generateSourceCode(symbolType)

		moduleDir := "testdata/test"
		abs, err := filepath.Abs(moduleDir)
		assert.NoError(t, err)
		filePath := filepath.Join(abs, "test.go")
		err = os.WriteFile(filePath, []byte(code), 0600)
		assert.NoError(t, err)
		defer os.Remove(abs)

		r, err := Extract(abs)
		assert.NoError(t, err)
		expected := tmpl(symbolType, typenames[symbolType])

		schema.SortModuleDecls(r.Module)
		assert.Equal(t, normaliseString(expected), normaliseString(r.Module.String()))
	})
}

func tmpl(symbolType string, typename string) string {
	var typeEnum string
	if symbolType != "any" {
		typeEnum = fmt.Sprintf(
			`
  enum Discriminator {
    Variant %s
  }

  export enum ExportedDiscriminator {
    ExportedVariant %s
  }
`, typename, typename)
	}

	var valueEnum string
	switch symbolType {
	case "int":
		valueEnum = `
enum Color: Int {
    Blue = 1
    Green = 2
    Red = 0
  }
`
	case "string":
		valueEnum = `
enum Color: String {
    Blue = "Blue"
    Green = "Green"
    Red = "Red"
  }
`
	}

	data := struct {
		TypeName  string
		TypeEnum  string
		ValueEnum string
	}{
		TypeName:  typename,
		TypeEnum:  typeEnum,
		ValueEnum: valueEnum,
	}

	const tmpl = `
module test {
  config cfg {{.TypeName}}
  secret secret {{.TypeName}}

  database postgres testDb

  export topic exported_topic {{.TypeName}}
  topic topic {{.TypeName}}
  subscription subscription test.topic

  typealias Alias {{.TypeName}}

  typealias EqualAlias {{.TypeName}}

  export typealias ExportedAlias {{.TypeName}}

  export typealias ExportedEqualAlias {{.TypeName}}
{{.ValueEnum}}{{.TypeEnum}}
  data Data {
    dataField {{.TypeName}}
    paramDataField test.ParameterizedData<{{.TypeName}}>
  }

  export data ExportedData {
    field String
  }

  data ParameterizedData<T> {
    field T
  }

  verb dataFunc(test.Data) test.Data

  verb emptyVerbFunc(Unit) Unit

  export verb exportedFunc({{.TypeName}}) {{.TypeName}}

  verb func({{.TypeName}}) {{.TypeName}}

  verb sinkFunc({{.TypeName}}) Unit

  verb sourceFunc(Unit) {{.TypeName}}
}
`

	t, err := template.New("test").Parse(tmpl)
	if err != nil {
		panic(err)
	}

	var result bytes.Buffer
	err = t.Execute(&result, data)
	if err != nil {
		panic(err)
	}

	return result.String()
}
