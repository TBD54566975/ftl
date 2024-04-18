//go:build integration

package cronjobs

import (
	"context"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	db "github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
	"github.com/benbjohnson/clock"
)

func TestServiceWithRealDal(t *testing.T) {
	t.Parallel()
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	conn := sqltest.OpenForTesting(ctx, t)
	dal, err := db.New(ctx, conn)
	assert.NoError(t, err)

	// Using a real clock because real db queries use db clock
	// delay until we are on an odd second
	clk := clock.New()
	if clk.Now().Second()%2 == 0 {
		time.Sleep(time.Second - time.Duration(clk.Now().Nanosecond())*time.Nanosecond)
	} else {
		time.Sleep(2*time.Second - time.Duration(clk.Now().Nanosecond())*time.Nanosecond)
	}

	testServiceWithDal(ctx, t, dal, clk)
}
