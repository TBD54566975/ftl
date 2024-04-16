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

func TestServiceWithDB(t *testing.T) {
	t.Parallel()
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	conn := sqltest.OpenForTesting(ctx, t)
	dal, err := db.New(ctx, conn)
	assert.NoError(t, err)

	clk := clock.New()

	verbCallCount := map[string]int{}
	verbCallCountLock := sync.Mutex{}

	// Using a real clock because real db queries use db clock
	// delay until we are on an odd second
	if clk.Now().Second()%2 == 0 {
		time.Sleep(time.Second - time.Duration(clk.Now().Nanosecond())*time.Nanosecond)
	} else {
		time.Sleep(2*time.Second - time.Duration(clk.Now().Nanosecond())*time.Nanosecond)
	}

	moduleName := "initial"
	jobsToCreate := newJobs(t, moduleName, "*/2 * * * * * *", clk, 20)

	deploymentKey, err := dal.CreateDeployment(ctx, "go", &schema.Module{
		Name: moduleName,
	}, []db.DeploymentArtefact{}, []db.IngressRoutingEntry{}, jobsToCreate)
	assert.NoError(t, err)

	err = dal.ReplaceDeployment(ctx, deploymentKey, 1)
	assert.NoError(t, err)

	_ = newControllers(ctx, 5, dal, func() clock.Clock { return clk }, func(ctx context.Context, r *connect.Request[ftlv1.CallRequest], o optional.Option[model.RequestKey], s string) (*connect.Response[ftlv1.CallResponse], error) {
		verbRef := schema.RefFromProto(r.Msg.Verb)

		verbCallCountLock.Lock()
		verbCallCount[verbRef.Name]++
		verbCallCountLock.Unlock()

		return &connect.Response[ftlv1.CallResponse]{}, nil
	})

	time.Sleep(time.Second * 2 * 3)

	for _, j := range jobsToCreate {
		count := verbCallCount[j.Verb.Name]
		assert.Equal(t, count, 3, "expected verb %s to be called 3 times", j.Verb.Name)
	}
}
