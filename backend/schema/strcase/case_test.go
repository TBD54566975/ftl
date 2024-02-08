package strcase

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestCamelCase(t *testing.T) {
	for _, tt := range []struct {
		input    string
		expected []string
	}{
		{"lowercase", []string{"lowercase"}},
		{"Class", []string{"Class"}},
		{"MyClass", []string{"My", "Class"}},
		{"MyC", []string{"My", "C"}},
		{"HTML", []string{"HTML"}},
		{"PDFLoader", []string{"PDF", "Loader"}},
		{"AString", []string{"A", "String"}},
		{"SimpleXMLParser", []string{"Simple", "XML", "Parser"}},
		{"vimRPCPlugin", []string{"vim", "RPC", "Plugin"}},
		{"GL11Version", []string{"GL", "11", "Version"}},
		{"99Bottles", []string{"99", "Bottles"}},
		{"May5", []string{"May", "5"}},
		{"BFG9000", []string{"BFG", "9000"}},
		{"BöseÜberraschung", []string{"Böse", "Überraschung"}},
		{"Two  spaces", []string{"Two", "  ", "spaces"}},
		{"BadUTF8\xe2\xe2\xa1", []string{"BadUTF8\xe2\xe2\xa1"}},
		{"snake_case", []string{"snake", "_", "case"}},
	} {
		actual := split(tt.input)
		assert.Equal(t, tt.expected, actual, "camelCase(%q) = %v; want %v", tt.input, actual, tt.expected)
	}
}

func TestLowerCamelCase(t *testing.T) {
	for _, tt := range []struct {
		input    string
		expected string
	}{
		{"lowercase", "lowercase"},
		{"Class", "class"},
		{"MyClass", "myClass"},
		{"MyC", "myC"},
		{"HTML", "html"},
		{"PDFLoader", "pdfLoader"},
		{"AString", "aString"},
		{"SimpleXMLParser", "simpleXmlParser"},
		{"vimRPCPlugin", "vimRpcPlugin"},
		{"GL11Version", "gl11Version"},
		{"99Bottles", "99Bottles"},
		{"May5", "may5"},
		{"BFG9000", "bfg9000"},
		{"BöseÜberraschung", "böseÜberraschung"},
		{"snake_case", "snake_Case"},
	} {
		actual := ToLowerCamel(tt.input)
		assert.Equal(t, tt.expected, actual, "LowerCamelCase(%q) = %v; want %v", tt.input, actual, tt.expected)
	}
}

func TestUpperCamelCase(t *testing.T) {
	for _, tt := range []struct {
		input    string
		expected string
	}{
		{"lowercase", "Lowercase"},
		{"Class", "Class"},
		{"MyClass", "MyClass"},
		{"MyC", "MyC"},
		{"HTML", "Html"},
		{"PDFLoader", "PdfLoader"},
		{"AString", "AString"},
		{"SimpleXMLParser", "SimpleXmlParser"},
		{"vimRPCPlugin", "VimRpcPlugin"},
		{"GL11Version", "Gl11Version"},
		{"99Bottles", "99Bottles"},
		{"May5", "May5"},
		{"BFG9000", "Bfg9000"},
		{"BöseÜberraschung", "BöseÜberraschung"},
		{"snake_case", "Snake_Case"},
	} {
		actual := ToUpperCamel(tt.input)
		assert.Equal(t, tt.expected, actual, "UpperCamelCase(%q) = %v; want %v", tt.input, actual, tt.expected)
	}
}

func TestLowerSnake(t *testing.T) {
	for _, tt := range []struct {
		input    string
		expected string
	}{
		{"lowercase", "lowercase"},
		{"Class", "class"},
		{"MyClass", "my_class"},
		{"MyC", "my_c"},
		{"HTML", "html"},
		{"PDFLoader", "pdf_loader"},
		{"AString", "a_string"},
		{"SimpleXMLParser", "simple_xml_parser"},
		{"vimRPCPlugin", "vim_rpc_plugin"},
		{"GL11Version", "gl_11_version"},
		{"99Bottles", "99_bottles"},
		{"May5", "may_5"},
		{"BFG9000", "bfg_9000"},
		{"BöseÜberraschung", "böse_überraschung"},
		{"snake_case", "snake_case"},
	} {
		actual := ToLowerSnake(tt.input)
		assert.Equal(t, tt.expected, actual, "LowerSnakeCase(%q) = %v; want %v", tt.input, actual, tt.expected)
	}
}

func TestUpperSnake(t *testing.T) {
	for _, tt := range []struct {
		input    string
		expected string
	}{
		{"lowercase", "LOWERCASE"},
		{"Class", "CLASS"},
		{"MyClass", "MY_CLASS"},
		{"MyC", "MY_C"},
		{"HTML", "HTML"},
		{"PDFLoader", "PDF_LOADER"},
		{"AString", "A_STRING"},
		{"SimpleXMLParser", "SIMPLE_XML_PARSER"},
		{"vimRPCPlugin", "VIM_RPC_PLUGIN"},
		{"GL11Version", "GL_11_VERSION"},
		{"99Bottles", "99_BOTTLES"},
		{"May5", "MAY_5"},
		{"BFG9000", "BFG_9000"},
		{"BöseÜberraschung", "BÖSE_ÜBERRASCHUNG"},
		{"snake_case", "SNAKE_CASE"},
	} {
		actual := ToUpperSnake(tt.input)
		assert.Equal(t, tt.expected, actual, "UpperSnakeCase(%q) = %v; want %v", tt.input, actual, tt.expected)
	}
}

func TestLowerKebabCase(t *testing.T) {
	for _, tt := range []struct {
		input    string
		expected string
	}{
		{"lowercase", "lowercase"},
		{"Class", "class"},
		{"MyClass", "my-class"},
		{"MyC", "my-c"},
		{"HTML", "html"},
		{"PDFLoader", "pdf-loader"},
		{"AString", "a-string"},
		{"SimpleXMLParser", "simple-xml-parser"},
		{"vimRPCPlugin", "vim-rpc-plugin"},
		{"GL11Version", "gl-11-version"},
		{"99Bottles", "99-bottles"},
		{"May5", "may-5"},
		{"BFG9000", "bfg-9000"},
		{"BöseÜberraschung", "böse-überraschung"},
		{"snake_case", "snake-case"},
	} {
		actual := ToLowerKebab(tt.input)
		assert.Equal(t, tt.expected, actual, "LowerKebabCase(%q) = %v; want %v", tt.input, actual, tt.expected)
	}
}

func TestUpperKebabCase(t *testing.T) {
	for _, tt := range []struct {
		input    string
		expected string
	}{
		{"lowercase", "LOWERCASE"},
		{"Class", "CLASS"},
		{"MyClass", "MY-CLASS"},
		{"MyC", "MY-C"},
		{"HTML", "HTML"},
		{"PDFLoader", "PDF-LOADER"},
		{"AString", "A-STRING"},
		{"SimpleXMLParser", "SIMPLE-XML-PARSER"},
		{"vimRPCPlugin", "VIM-RPC-PLUGIN"},
		{"GL11Version", "GL-11-VERSION"},
		{"99Bottles", "99-BOTTLES"},
		{"May5", "MAY-5"},
		{"BFG9000", "BFG-9000"},
		{"BöseÜberraschung", "BÖSE-ÜBERRASCHUNG"},
		{"snake_case", "SNAKE-CASE"},
	} {
		actual := ToUpperKebab(tt.input)
		assert.Equal(t, tt.expected, actual, "UpperKebabCase(%q) = %v; want %v", tt.input, actual, tt.expected)
	}
}
