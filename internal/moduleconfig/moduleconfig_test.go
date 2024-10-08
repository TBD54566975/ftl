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
				DeployDir:          "deploydir",
				GeneratedSchemaDir: "generatedschemadir",
				Watch:              []string{"a", "b", "c"},
			},
			expected: ModuleConfig{
				Realm:              "home",
				Dir:                "a",
				Module:             "nothingset",
				Language:           "test",
				Build:              "build",
				DeployDir:          "deploydir",
				GeneratedSchemaDir: "generatedschemadir",
				Watch:              []string{"a", "b", "c"},
			},
		},
		{
			config: UnvalidatedModuleConfig{
				Dir:                "b",
				Module:             "allset",
				Language:           "test",
				Build:              "custombuild",
				DeployDir:          "customdeploydir",
				GeneratedSchemaDir: "customgeneratedschemadir",
				Watch:              []string{"custom1"},
				LanguageConfig: map[string]any{
					"build-tool": "maven",
					"more":       []int{1, 2, 3},
				},
			},
			defaults: CustomDefaults{
				Build:              "build",
				DeployDir:          "deploydir",
				GeneratedSchemaDir: "generatedschemadir",
				Watch:              []string{"a", "b", "c"},
			},
			expected: ModuleConfig{
				Realm:              "home",
				Dir:                "b",
				Module:             "allset",
				Language:           "test",
				Build:              "custombuild",
				DeployDir:          "customdeploydir",
				GeneratedSchemaDir: "customgeneratedschemadir",
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
				DeployDir: "deploydir",
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
				Module:   "nodeploydir",
				Language: "test",
			},
			defaults: CustomDefaults{},
			error:    "no deploy directory configured",
		},
	} {
		t.Run(tt.config.Module, func(t *testing.T) {
			t.Parallel()

			config, err := tt.config.FillDefaultsAndValidate(tt.defaults)
			if tt.error != "" {
				assert.EqualError(t, err, tt.error)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, config)
		})
	}
}
