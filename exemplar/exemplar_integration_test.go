//go:build integration

package exemplar

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	in "github.com/TBD54566975/ftl/internal/integration"
	"github.com/alecthomas/assert/v2"
)

func TestExemplar(t *testing.T) {
	logFilePath := filepath.Join(t.TempDir(), "exemplar.log")
	var postResult struct {
		ID int `json:"id"`
	}
	var successAgentId int
	var failedAgentId int
	in.Run(t,
		// in.WithFTLConfig("../../../ftl-project.toml"),
		in.WithTestDataDir("."),
		in.CopyModule("origin"),
		in.CopyModule("relay"),
		in.CopyModule("pulse"),
		in.CreateDBAction("relay", "exemplardb", false),
		in.Deploy("origin"),
		in.Deploy("relay"),
		in.Deploy("pulse"),

		// TODO abcd
		in.ExecWithOutput("ftl", []string{"config", "set", "origin.nonce", "--inline", "abcd"}, func(output string) {
			fmt.Println(output)
		}),

		in.ExecWithOutput("ftl", []string{"config", "set", "relay.log_file", "--inline", logFilePath}, func(output string) {
			fmt.Println(output)
		}),

		in.ExecWithOutput("ftl", []string{"config", "set", "pulse.log_file", "--inline", logFilePath}, func(output string) {
			fmt.Println(output)
		}),

		in.ExecWithOutput("curl", []string{"-s", "-X", "POST", "http://127.0.0.1:8891/http/agent", "-H", "Content-Type: application/json", "-d", `{"alias": "bond", "license_to_kill": true, "hired_at": "2023-10-23T23:20:45.00Z"}`}, func(output string) {
			err := json.Unmarshal([]byte(output), &postResult)
			assert.NoError(t, err)
			successAgentId = postResult.ID
			fmt.Printf("successAgentId: %d\n", successAgentId)
		}),

		in.ExecWithOutput("curl", []string{"-s", "-X", "POST", "http://127.0.0.1:8891/http/agent", "-H", "Content-Type: application/json", "-d", `{"alias": "notbond", "license_to_kill": false, "hired_at": "2024-08-12T21:10:37.00Z"}`}, func(output string) {
			err := json.Unmarshal([]byte(output), &postResult)
			assert.NoError(t, err)
			failedAgentId = postResult.ID
			fmt.Printf("failedAgentId: %d\n", failedAgentId)
		}),

		in.Sleep(2*1000*1000),

		in.ExecWithOutput("ftl", []string{"call", "relay.missionResult", fmt.Sprintf(`{"agent_id": %d, "successful": true}`, successAgentId)}, func(output string) {
			fmt.Println(output)
		}),

		in.ExecWithOutput("ftl", []string{"call", "relay.missionResult", fmt.Sprintf(`{"agent_id": %d, "successful": false}`, failedAgentId)}, func(output string) {
			fmt.Println(output)
		}),

		in.Sleep(2*1000*1000),

		in.FileContains(logFilePath, fmt.Sprintf("deployed %d", successAgentId)),
		in.FileContains(logFilePath, fmt.Sprintf("deployed %d", failedAgentId)),
		in.FileContains(logFilePath, fmt.Sprintf("succeeded %d", successAgentId)),
		in.FileContains(logFilePath, fmt.Sprintf("terminated %d", failedAgentId)),
	)
}
