package localdebug

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/TBD54566975/ftl/internal/log"
)

const jvmDebugConfig = `<component name="ProjectRunConfigurationManager">
  <configuration default="false" name="%s" type="Remote">
    <option name="USE_SOCKET_TRANSPORT" value="true" />
    <option name="SERVER_MODE" value="false" />
    <option name="SHMEM_ADDRESS" />
    <option name="HOST" value="localhost" />
    <option name="PORT" value="%d" />
    <option name="AUTO_RESTART" value="false" />
    <RunnerSettings RunnerId="Debug">
      <option name="DEBUG_PORT" value="%d" />
      <option name="LOCAL" value="false" />
    </RunnerSettings>
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

func init() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		lock.Lock()
		defer lock.Unlock()
		for p := range paths {
			println(p)
			_ = os.Remove(p)
		}

	}()
}

var lock = sync.Mutex{}
var paths = map[string]bool{}

// This is a bit of a hack to prove out the concept of local debugging.
func SyncIDEDebugIntegrations(cxt context.Context, projectPath string, ports map[string]*DebugInfo) {
	lock.Lock()
	defer lock.Unlock()
	if projectPath == "" {
		return
	}
	handleIntellij(cxt, projectPath, ports)
}

func handleIntellij(cxt context.Context, path string, ports map[string]*DebugInfo) {
	logger := log.FromContext(cxt)
	ideaPath := findIdeaFolder(path)
	if ideaPath == "" {
		return
	}
	runConfig := filepath.Join(ideaPath, "runConfigurations")
	err := os.MkdirAll(runConfig, 0750)
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
			if debugInfo != nil {
				continue
			} else {
				// old FTL entry, remove it
				path := filepath.Join(runConfig, entry.Name())
				_ = os.Remove(path)
			}
		}
	}
	for k, v := range ports {
		if v.Language == "java" || v.Language == "kotlin" {
			name := filepath.Join(runConfig, "FTL."+k+".xml")
			err := os.WriteFile(name, []byte(fmt.Sprintf(jvmDebugConfig, "FT𝝺 JVM - "+k, v.Port, v.Port)), 0644)
			paths[name] = true
			if err != nil {
				logger.Errorf(err, "could not create FTL Java Config")
				return
			}
		} else if v.Language == "go" {
			name := filepath.Join(runConfig, "FTL."+k+".xml")
			err := os.WriteFile(name, []byte(fmt.Sprintf(golangDebugConfig, "FT𝝺 GO - "+k, v.Port)), 0644)
			paths[name] = true
			if err != nil {
				logger.Errorf(err, "could not create FTL Go Config")
				return
			}
		}
	}
}

// findIdeaFolder recurses up the directory tree to find a .idea folder.
func findIdeaFolder(startPath string) string {
	currentPath := startPath

	for {
		ideaPath := filepath.Join(currentPath, ".idea")
		if _, err := os.Stat(ideaPath); err == nil {
			return ideaPath
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

type DebugInfo struct {
	Port     int
	Language string
}
