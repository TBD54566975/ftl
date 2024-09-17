package localdebug

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TBD54566975/ftl/internal/log"
)

const projectToml = "project.toml"
const jvmDebugConfig = `<component name="ProjectRunConfigurationManager">
  <configuration default="false" name="%s" type="Remote">
    <option name="USE_SOCKET_TRANSPORT" value="true" />
    <option name="SERVER_MODE" value="false" />
    <option name="SHMEM_ADDRESS" />
    <option name="HOST" value="localhost" />
    <option name="PORT" value="%d" />
    <option name="AUTO_RESTART" value="false" />
    <method v="2" />
  </configuration>
</component>`

const golangDebugConfig = `<component name="ProjectRunConfigurationManager">
  <configuration default="false" name="%s" type="GoRemoteDebugConfigurationType" factoryName="Go Remote" port="%d">
    <option name="disconnectOption" value="LEAVE" />
    <disconnect value="LEAVE" />
    <method v="2" />
  </configuration>
</component>`

type IDEIntegration struct {
	projectPath string
}

type DebugInfo struct {
	Port     int
	Language string
}

func NewIDEIntegration(projectPath string) *IDEIntegration {
	ret := &IDEIntegration{projectPath: projectPath}
	return ret
}

// SyncIDEDebugIntegrations will sync the local IDE debug configurations for the given project path.
// This is a bit of a hack to prove out the concept of local debugging.
func (r *IDEIntegration) SyncIDEDebugIntegrations(ctx context.Context, ports map[string]*DebugInfo) {
	if r.projectPath == "" {
		return
	}
	r.handleIntellij(ctx, ports)
	r.handleVSCode(ctx, ports)
}

// GetExistingDebugPort will return the existing debug ports for the given project name
// based on VSCode configurations.
// VSCode configurations are always created, so there is no need to look at intellij
func (r *IDEIntegration) GetExistingDebugPort(ctx context.Context, module string) int {
	if r.projectPath == "" {
		return 0
	}
	logger := log.FromContext(ctx)
	vscode := r.findFolder(".vscode", true)
	launchJSON := filepath.Join(vscode, "launch.json")
	contents := map[string]any{}
	var configurations []any
	if _, err := os.Stat(launchJSON); err == nil {
		file, err := os.ReadFile(launchJSON)
		if err != nil {
			logger.Errorf(err, "could not read launch.json")
			return 0
		}
		err = json.Unmarshal(file, &contents)
		if err != nil {
			logger.Errorf(err, "could not read launch.json")
			return 0
		}
		configurations = contents["configurations"].([]any) //nolint:forcetypeassert
		if configurations == nil {
			configurations = []any{}
		}
		for _, config := range configurations {
			name := config.(map[string]any)["name"].(string) //nolint:forcetypeassert
			if strings.HasPrefix(name, "FT𝝺") && strings.HasSuffix(name, "- "+module) {
				ret, ok := config.(map[string]any)["port"].(float64)
				if !ok {
					logger.Warnf("could not read port from launch.json")
					return 0
				}
				return int(ret)
			}
		}
	}
	return 0
}

func (r *IDEIntegration) handleIntellij(ctx context.Context, ports map[string]*DebugInfo) {
	logger := log.FromContext(ctx)
	ideaPath := r.findFolder(".idea", false)
	if ideaPath == "" {
		return
	}
	runConfig := filepath.Join(ideaPath, "runConfigurations")
	err := os.MkdirAll(runConfig, 0600)
	if err != nil {
		logger.Errorf(err, "could not create runConfigurations directory")
		return
	}
	entries, err := os.ReadDir(runConfig)
	if err != nil {
		logger.Errorf(err, "could not create runConfigurations directory")
		return
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "FTL.") && strings.HasSuffix(entry.Name(), ".xml") {
			debugInfo := ports[entry.Name()[4:(len(entry.Name())-4)]]
			if debugInfo == nil {
				// old FTL entry, remove it
				path := filepath.Join(runConfig, entry.Name())
				_ = os.Remove(path)
			}
		}
	}
	for k, v := range ports {
		if v.Language == "java" || v.Language == "kotlin" {
			name := filepath.Join(runConfig, "FTL."+k+".xml")
			err := os.WriteFile(name, []byte(fmt.Sprintf(jvmDebugConfig, "FT𝝺 JVM - "+k, v.Port)), 0600)
			if err != nil {
				logger.Errorf(err, "could not create FTL Java Config")
				return
			}
		} else if v.Language == "go" {
			name := filepath.Join(runConfig, "FTL."+k+".xml")
			err := os.WriteFile(name, []byte(fmt.Sprintf(golangDebugConfig, "FT𝝺 GO - "+k, v.Port)), 0600)
			if err != nil {
				logger.Errorf(err, "could not create FTL Go Config")
				return
			}
		}
	}
}

// findFolder recurses up the directory tree to find a folder
// If it can't find one it returns the path that would exist next to project.toml
func (r *IDEIntegration) findFolder(folder string, allowNonExistent bool) string {
	currentPath := r.projectPath

	for {
		searchPath := filepath.Join(currentPath, folder)
		if _, err := os.Stat(searchPath); err == nil {
			return searchPath
		}
		projectPath := filepath.Join(currentPath, projectToml)
		if _, err := os.Stat(projectPath); err == nil {
			// Reached the project.toml file, we don't go outside of the project
			if allowNonExistent {
				return searchPath
			}
			return ""
		}
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached the root directory
			break
		}
		currentPath = parentPath
	}
	return ""
}

func (r *IDEIntegration) handleVSCode(ctx context.Context, ports map[string]*DebugInfo) {
	logger := log.FromContext(ctx)
	vscode := r.findFolder(".vscode", true)
	err := os.MkdirAll(vscode, 0600)
	if err != nil {
		logger.Errorf(err, "could not create .vscode directory")
		return
	}
	launchJSON := filepath.Join(vscode, "launch.json")

	contents := map[string]any{}
	existing := map[string]int{}
	var configurations []any
	if _, err := os.Stat(launchJSON); err == nil {
		file, err := os.ReadFile(launchJSON)
		if err != nil {
			logger.Errorf(err, "could not read launch.json")
			return
		}
		err = json.Unmarshal(file, &contents)
		if err != nil {
			logger.Errorf(err, "could not read launch.json")
			return
		}
		configurations = contents["configurations"].([]any) //nolint:forcetypeassert
		if configurations == nil {
			configurations = []any{}
		}
	} else {
		contents["version"] = "0.2.0"
		configurations = []any{}
	}
	for i, config := range configurations {
		name := config.(map[string]any)["name"].(string) //nolint:forcetypeassert
		if strings.HasPrefix(name, "FT𝝺") {
			existing[name] = i
		}
	}

	for k, v := range ports {
		if v.Language == "java" || v.Language == "kotlin" {
			name := "FT𝝺 JVM - " + k
			pos, ok := existing[name]
			delete(existing, name)
			if ok {
				// Update the port
				configurations[pos].(map[string]any)["port"] = v.Port //nolint:forcetypeassert
				continue
			}
			entry := map[string]any{
				"name":     name,
				"type":     "java",
				"request":  "attach",
				"hostName": "127.0.0.1",
				"port":     v.Port,
			}
			configurations = append(configurations, entry)

		} else if v.Language == "go" {
			name := "FT𝝺 GO - " + k
			pos, ok := existing[name]
			if ok {
				// Update the port
				configurations[pos].(map[string]any)["port"] = v.Port //nolint:forcetypeassert
				continue
			}
			entry := map[string]any{
				"name":       name,
				"type":       "go",
				"request":    "attach",
				"mode":       "remote",
				"apiVersion": 2,
				"host":       "127.0.0.1",
				"port":       v.Port,
			}
			configurations = append(configurations, entry)
		}
	}
	contents["configurations"] = configurations
	data, err := json.MarshalIndent(contents, "", "  ")
	if err != nil {
		logger.Errorf(err, "could not marshal launch.json")
		return
	}
	err = os.WriteFile(launchJSON, data, 0600)
	if err != nil {
		logger.Errorf(err, "could not write launch.json")
		return
	}
}
