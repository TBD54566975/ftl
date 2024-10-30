//go:build integration

package smoketest

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	in "github.com/TBD54566975/ftl/internal/integration"
)

func TestExemplar(t *testing.T) {
	tmpDir := t.TempDir()
	logFilePath := filepath.Join(tmpDir, "smoketest.log")

	var postResult struct {
		ID int `json:"id"`
	}
	var successAgentId int = 7
	var failedAgentId int = 99
	nonce := randomString(4)

	skipKubeFullDeploy := os.Getenv("SKIP_KUBE_FULL_DEPLOY") == "true"
	fmt.Println("skipKubeFullDeploy:", skipKubeFullDeploy)

	in.Run(t,
		in.WithKubernetes(!skipKubeFullDeploy),
		in.WithJavaBuild(),
		in.WithFTLConfig("../../../ftl-project.toml"),
		in.WithTestDataDir("."),
		in.CopyModule("origin"),
		in.CopyModule("relay"),
		in.CopyModule("pulse"),
		in.CreateDBAction("relay", "exemplardb", false),

		in.ExecWithOutput("ftl", []string{"config", "set", "origin.nonce", "--inline", nonce}, func(output string) {
			fmt.Println(output)
		}),

		in.ExecWithOutput("ftl", []string{"config", "set", "relay.log_file", "--inline", logFilePath}, func(output string) {
			fmt.Println(output)
		}),

		in.Deploy("origin"),
		in.Deploy("relay"),
		in.Deploy("pulse"),

		in.ExecWithOutput("curl", []string{"-s", "-X", "POST", "http://127.0.0.1:8891/http/agent", "-H", "Content-Type: application/json", "-d", fmt.Sprintf(`{"id": %v, "alias": "james", "license_to_kill": true, "hired_at": "2023-10-23T23:20:45.00Z"}`, successAgentId)}, func(output string) {
			err := json.Unmarshal([]byte(output), &postResult)
			assert.NoError(t, err)
			assert.Equal(t, successAgentId, postResult.ID)
		}),

		in.ExecWithOutput("curl", []string{"-s", "-X", "POST", "http://127.0.0.1:8891/http/agent", "-H", "Content-Type: application/json", "-d", fmt.Sprintf(`{"id": %v, "alias": "bill", "license_to_kill": false, "hired_at": "2024-08-12T21:10:37.00Z"}`, failedAgentId)}, func(output string) {
			err := json.Unmarshal([]byte(output), &postResult)
			assert.NoError(t, err)
			assert.Equal(t, failedAgentId, postResult.ID)
		}),

		in.Sleep(2*time.Second),

		in.ExecWithOutput("ftl", []string{"call", "relay.missionResult", fmt.Sprintf(`{"agentId": %v, "successful": true}`, successAgentId)}, func(output string) {
			fmt.Println(output)
		}),

		in.ExecWithOutput("ftl", []string{"call", "relay.missionResult", fmt.Sprintf(`{"agentId": %v, "successful": false}`, failedAgentId)}, func(output string) {
			fmt.Println(output)
		}),

		in.Sleep(2*time.Second),

		in.FileContains(logFilePath, fmt.Sprintf("deployed %d", successAgentId)),
		in.FileContains(logFilePath, fmt.Sprintf("deployed %d", failedAgentId)),
		in.FileContains(logFilePath, fmt.Sprintf("succeeded %d", successAgentId)),
		in.FileContains(logFilePath, fmt.Sprintf("terminated %d", failedAgentId)),
		in.FileContains(logFilePath, fmt.Sprintf("cron %v", nonce)),
	)
}

const charset = "abcdefghijklmnopqrstuvwxyz"

func randomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
