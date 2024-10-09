//go:build integration

package smoketest

import (
	"fmt"
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"

	in "github.com/TBD54566975/ftl/internal/integration"
)

func TestSmokeEcho(t *testing.T) {
	skipKubeFullDeploy := os.Getenv("SKIP_KUBE_FULL_DEPLOY") == "true"
	fmt.Println("skipKubeFullDeploy:", skipKubeFullDeploy)
	in.Run(t,
		in.WithKubernetes(!skipKubeFullDeploy),
		in.WithTestDataDir("."),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.Call("echo", "echo", "Joe", func(t testing.TB, response string) {
			expected := fmt.Sprintf("Hello, %s!!!", "Joe")
			assert.Equal(t, expected, response)
		}),
		in.Exec("ftl", "--version"),
	)
}
