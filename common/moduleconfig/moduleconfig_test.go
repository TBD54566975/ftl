package moduleconfig

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseImportsFromTestData(t *testing.T) {
	testFilePath := filepath.Join("testdata", "imports.go")
	expectedImports := []string{"fmt", "os"}
	imports, err := parseImports(testFilePath)
	if err != nil {
		t.Fatalf("Failed to parse imports: %v", err)
	}

	if !reflect.DeepEqual(imports, expectedImports) {
		t.Errorf("parseImports() got = %v, want %v", imports, expectedImports)
	}
}
