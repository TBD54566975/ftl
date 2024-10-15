//go:build integration

package cronjobs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	in "github.com/TBD54566975/ftl/internal/integration"
)

func TestCron(t *testing.T) {
	dir := t.TempDir()
	// Due to some MacOS magic, /tmp differs between this test code and the
	// executing module, so we need to pass the file path as an environment
	// variable.
	tmpFile := filepath.Join(dir, "cron.txt")
	t.Setenv("DEST_FILE", tmpFile)

	t.Cleanup(func() { _ = os.Remove(tmpFile) })

	in.Run(t,
		in.WithLanguages("go", "java"),
		in.CopyModule("cron"),
		in.Deploy("cron"),
		in.Wait("cron"),
		in.Sleep(1*time.Second),
		func(t testing.TB, ic in.TestContext) {
			_, err := os.Stat(tmpFile)
			assert.NoError(t, err)
		},
	)
}

func TestCronIsRemoved(t *testing.T) {
	dir := t.TempDir()
	// We want to make sure that cron jobs are shut down when the deployment is updated
	// And we don't end up with double invocations
	// To test this we are going to remove the cron and turn it into a normal verb
	// If the verb is still invoked after the redeploy then we have a problem
	tmpFile := filepath.Join(dir, "cron.txt")
	t.Setenv("DEST_FILE", tmpFile)

	t.Cleanup(func() { _ = os.Remove(tmpFile) })

	in.Run(t,
		in.WithLanguages("go"),
		in.CopyModule("cron"),
		in.Deploy("cron"),
		in.Wait("cron"),
		in.Sleep(1*time.Second),
		func(t testing.TB, ic in.TestContext) {
			_, err := os.Stat(tmpFile)
			assert.NoError(t, err)
			data, err := os.ReadFile(tmpFile)
			assert.NoError(t, err)
			assert.True(t, strings.Contains(string(data), "Hello, world!"))
		},
		in.EditFile("cron/cron.go", func(content []byte) []byte {
			ret := strings.ReplaceAll(string(content), "//ftl:cron * * * * * * *", "//ftl:verb")
			ret = strings.ReplaceAll(ret, "Hello, world!", "NEW VERB")
			return []byte(ret)
		}),
		in.Deploy("cron"),
		func(t testing.TB, ic in.TestContext) {
			time.Sleep(2 * time.Second)
			data, err := os.ReadFile(tmpFile)
			assert.NoError(t, err)
			assert.False(t, strings.Contains(string(data), "NEW VERB"))
		},
	)
}
