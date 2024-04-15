package projectconfig

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestMerge(t *testing.T) {
	a := Config{
		Global: ConfigAndSecrets{
			Config: map[string]*URL{
				"key1": MustParseURL("inline://foo"),
			},
			Secrets: map[string]*URL{
				"key2": MustParseURL("inline://bar"),
			},
		},
		Modules: map[string]ConfigAndSecrets{
			"test": {
				Config: map[string]*URL{
					"key3": MustParseURL("inline://baz"),
				},
				Secrets: map[string]*URL{
					"key4":  MustParseURL("inline://qux"),
					"key10": MustParseURL("inline://qux10"),
				},
			},
		},
		ModuleDirs:   []string{"a/b/c"},
		ExternalDirs: []string{"e/f"},
		Executables: Executables{
			FTL: "ftl",
		},
		Commands: Commands{
			Startup: []string{"echo 'Before'"},
		},
		FTLMinVersion: "0.0.1",
	}
	b := Config{
		Global: ConfigAndSecrets{
			Config: map[string]*URL{
				"key1": MustParseURL("inline://foo2"),
				"key5": MustParseURL("inline://foo5"),
			},
			Secrets: map[string]*URL{
				"key2": MustParseURL("inline://bar2"),
				"key6": MustParseURL("inline://bar6"),
			},
		},
		Modules: map[string]ConfigAndSecrets{
			"test": {
				Config: map[string]*URL{
					"key3": MustParseURL("inline://baz2"),
					"key7": MustParseURL("inline://baz7"),
				},
				Secrets: map[string]*URL{
					"key4": MustParseURL("inline://qux2"),
					"key8": MustParseURL("inline://qux8"),
				},
			},
		},
		ModuleDirs:   []string{"d"},
		ExternalDirs: []string{"g/h"},
		Executables: Executables{
			FTL: "./bin/ftl",
		},
		Commands: Commands{
			Startup: []string{"echo 'After'"},
		},
		FTLMinVersion: "0.0.2",
	}
	a = merge(a, b)
	expected := Config{
		Global: ConfigAndSecrets{
			Config: map[string]*URL{
				"key1": MustParseURL("inline://foo2"),
				"key5": MustParseURL("inline://foo5"),
			},
			Secrets: map[string]*URL{
				"key2": MustParseURL("inline://bar2"),
				"key6": MustParseURL("inline://bar6"),
			},
		},
		Modules: map[string]ConfigAndSecrets{
			"test": {
				Config: map[string]*URL{
					"key3": MustParseURL("inline://baz2"),
					"key7": MustParseURL("inline://baz7"),
				},
				Secrets: map[string]*URL{
					"key4":  MustParseURL("inline://qux2"),
					"key8":  MustParseURL("inline://qux8"),
					"key10": MustParseURL("inline://qux10"),
				},
			},
		},
		ModuleDirs:   []string{"d"},
		ExternalDirs: []string{"g/h"},
		Executables: Executables{
			FTL: "./bin/ftl",
		},
		Commands: Commands{
			Startup: []string{"echo 'After'"},
		},
		FTLMinVersion: "0.0.2",
	}
	assert.Equal(t, expected, a)
}
