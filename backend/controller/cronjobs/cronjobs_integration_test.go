//go:build integration

package cronjobs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/benbjohnson/clock"

	db "github.com/TBD54566975/ftl/backend/controller/cronjobs/dal"
	parentdb "github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	"github.com/TBD54566975/ftl/internal/encryption"
	in "github.com/TBD54566975/ftl/internal/integration"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestServiceWithRealDal(t *testing.T) {
	t.Parallel()
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	conn := sqltest.OpenForTesting(ctx, t)
	dal := db.New(conn)
	parentDAL, err := parentdb.New(ctx, conn, encryption.NewBuilder())
	assert.NoError(t, err)

	// Using a real clock because real db queries use db clock
	// delay until we are on an odd second
	clk := clock.New()
	if clk.Now().Second()%2 == 0 {
		time.Sleep(time.Second - time.Duration(clk.Now().Nanosecond())*time.Nanosecond)
	} else {
		time.Sleep(2*time.Second - time.Duration(clk.Now().Nanosecond())*time.Nanosecond)
	}

	testServiceWithDal(ctx, t, dal, parentDAL, clk)
}

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
		func(t testing.TB, ic in.TestContext) {
			_, err := os.Stat(tmpFile)
			assert.NoError(t, err)
		},
	)
}
