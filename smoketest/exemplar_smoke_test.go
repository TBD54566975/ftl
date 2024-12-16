//go:build smoketest

package smoketest

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	in "github.com/block/ftl/internal/integration"
)

func TestExemplarSmoke(t *testing.T) {
	tmpDir := "/tmp"
	logFilePath := filepath.Join(tmpDir, "smoketest.log")

	var postResult struct {
		ID int `json:"id"`
	}
	var successAgentId int = 7
	var failedAgentId int = 99
	nonce := randomString(4)

	in.Run(t,
		in.WithKubernetes(),
		in.WithJavaBuild(),
		in.WithFTLConfig("../../../ftl-project.toml"),
		in.WithTestDataDir("."),
		in.CopyModule("origin"),
		in.CopyModule("relay"),
		in.CopyModule("pulse"),

		in.ExecWithOutput("ftl", []string{"config", "set", "origin.nonce", "--db", nonce}, func(output string) {
			fmt.Println(output)
		}),

		in.ExecWithOutput("ftl", []string{"config", "set", "relay.log_file", "--db", logFilePath}, func(output string) {
			fmt.Println(output)
		}),

		in.Deploy("origin"),
		in.Deploy("relay"),
		in.Deploy("pulse"),

		in.ExecWithOutput("curl", []string{"-s", "-X", "POST", "http://127.0.0.1:8891/ingress/agent", "-H", "Content-Type: application/json", "-d", fmt.Sprintf(`{"id": %v, "alias": "james", "license_to_kill": true, "hired_at": "2023-10-23T23:20:45.00Z"}`, successAgentId)}, func(output string) {
			fmt.Printf("output: %s\n", output)
			err := json.Unmarshal([]byte(output), &postResult)
			assert.NoError(t, err)
			assert.Equal(t, successAgentId, postResult.ID)
		}),

		in.ExecWithOutput("curl", []string{"-s", "-X", "POST", "http://127.0.0.1:8891/ingress/agent", "-H", "Content-Type: application/json", "-d", fmt.Sprintf(`{"id": %v, "alias": "bill", "license_to_kill": false, "hired_at": "2024-08-12T21:10:37.00Z"}`, failedAgentId)}, func(output string) {
			fmt.Printf("output: %s\n", output)
			err := json.Unmarshal([]byte(output), &postResult)
			assert.NoError(t, err)
			assert.Equal(t, failedAgentId, postResult.ID)
		}),

		in.Sleep(5*time.Second),

		in.ExecWithOutput("ftl", []string{"call", "relay.missionResult", fmt.Sprintf(`{"agentId": %v, "successful": true}`, successAgentId)}, func(output string) {
			fmt.Println(output)
		}),

		in.ExecWithOutput("ftl", []string{"call", "relay.missionResult", fmt.Sprintf(`{"agentId": %v, "successful": false}`, failedAgentId)}, func(output string) {
			fmt.Println(output)
		}),

		in.Sleep(2*time.Second),

		in.Call("relay", "fetchLogs", in.Obj{}, func(t testing.TB, resp in.Obj) {
			fmt.Printf("fetchLogs: %v\n", resp)
		}),
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
