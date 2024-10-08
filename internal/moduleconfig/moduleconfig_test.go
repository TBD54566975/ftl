package moduleconfig

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestDefaulting(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		config   UnvalidatedModuleConfig
		defaults CustomDefaults
		expected ModuleConfig
		error    string
	}{
		{
			config: UnvalidatedModuleConfig{
				Dir:      "a",
				Module:   "nothingset",
				Language: "test",
			},
			defaults: CustomDefaults{
				Build:              "build",
				Deploy:             []string{"deploy"},
				DeployDir:          "deploydir",
				GeneratedSchemaDir: "generatedschemadir",
				Errors:             "errors.pb",
				Watch:              []string{"a", "b", "c"},
			},
			expected: ModuleConfig{
				Realm:              "home",
				Dir:                "a",
				Module:             "nothingset",
				Language:           "test",
				Build:              "build",
				Deploy:             []string{"deploy"},
				DeployDir:          "deploydir",
				GeneratedSchemaDir: "generatedschemadir",
				Errors:             "errors.pb",
				Watch:              []string{"a", "b", "c"},
			},
		},
		{
			config: UnvalidatedModuleConfig{
				Dir:                "b",
				Module:             "allset",
				Language:           "test",
				Build:              "custombuild",
				Deploy:             []string{"customdeploy"},
				DeployDir:          "customdeploydir",
				GeneratedSchemaDir: "customgeneratedschemadir",
				Errors:             "customerrors.pb",
				Watch:              []string{"custom1"},
				LanguageConfig: map[string]any{
					"build-tool": "maven",
					"more":       []int{1, 2, 3},
				},
			},
			defaults: CustomDefaults{
				Build:              "build",
				Deploy:             []string{"deploy"},
				DeployDir:          "deploydir",
				GeneratedSchemaDir: "generatedschemadir",
				Errors:             "errors.pb",
				Watch:              []string{"a", "b", "c"},
			},
			expected: ModuleConfig{
				Realm:              "home",
				Dir:                "b",
				Module:             "allset",
				Language:           "test",
				Build:              "custombuild",
				Deploy:             []string{"customdeploy"},
				DeployDir:          "customdeploydir",
				GeneratedSchemaDir: "customgeneratedschemadir",
				Errors:             "customerrors.pb",
				Watch:              []string{"custom1"},
				LanguageConfig: map[string]any{
					"build-tool": "maven",
					"more":       []int{1, 2, 3},
				},
			},
		},

		{
			config: UnvalidatedModuleConfig{
				Dir:      "b",
				Module:   "languageconfig",
				Language: "test",
				LanguageConfig: map[string]any{
					"alreadyset": "correct",
					"nodefault":  []int{1, 2, 3},
					"root": map[string]any{
						"nested1": "actualvalue1",
					},
				},
			},
			defaults: CustomDefaults{
				Deploy:    []string{"example"},
				DeployDir: "deploydir",
				LanguageConfig: map[string]any{
					"alreadyset": "incorrect",
					"notset":     "defaulted",
					"root": map[string]any{
						"nested1": "value1",
						"nested2": "value2",
					},
				},
			},
			expected: ModuleConfig{
				Deploy:    []string{"example"},
				DeployDir: "deploydir",
				Errors:    "errors.pb",
				Realm:     "home",
				Dir:       "b",
				Module:    "languageconfig",
				Language:  "test",
				LanguageConfig: map[string]any{
					"alreadyset": "correct",
					"nodefault":  []int{1, 2, 3},
					"root": map[string]any{
						"nested1": "actualvalue1",
					},
					"notset": "defaulted",
				},
			},
		},

		// Validation failures
		{
			config: UnvalidatedModuleConfig{
				Dir:      "b",
				Module:   "nodeploy",
				Language: "test",
			},
			defaults: CustomDefaults{
				DeployDir: "deploydir",
			},
			error: "no deploy files configured",
		},
		{
			config: UnvalidatedModuleConfig{
				Dir:      "b",
				Module:   "nodeploydir",
				Language: "test",
			},
			defaults: CustomDefaults{
				Deploy: []string{"example"},
			},
			error: "no deploy directory configured",
		},
		{
			config: UnvalidatedModuleConfig{
				Dir:      "b",
				Module:   "deploynotindir",
				Language: "test",
			},
			defaults: CustomDefaults{
				Deploy:    []string{"example"},
				DeployDir: "../../deploydir",
			},
			error: "must be relative to the module directory",
		},
	} {
		t.Run(tt.config.Module, func(t *testing.T) {
			t.Parallel()

			config, err := tt.config.FillDefaultsAndValidate(tt.defaults)
			if tt.error != "" {
				assert.Contains(t, err.Error(), tt.error)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, config)
		})
	}
}
