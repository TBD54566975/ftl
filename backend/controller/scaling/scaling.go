package scaling

import (
	"context"
	"net/url"
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/jpillora/backoff"

	"github.com/TBD54566975/ftl/backend/controller/leases"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type RunnerScaling func(ctx context.Context, endpoint url.URL, leaser leases.Leaser) error

func BeginGrpcScaling(ctx context.Context, url url.URL, leaser leases.Leaser, handler func(ctx context.Context, msg *ftlv1.PullSchemaResponse) error) {
	leaseTimeout := time.Second * 20
	for {
		// Grab a lease to take control of runner scaling
		lease, leaseContext, err := leaser.AcquireLease(ctx, leases.SystemKey("ftl-scaling", "runner-creation"), leaseTimeout, optional.None[any]())
		if err == nil {
			defer func(lease leases.Lease) {
				err := lease.Release()
				if err != nil {
					logger := log.FromContext(ctx)
					logger.Errorf(err, "Failed to release lease")
				}
			}(lease)
			// If we get it then we take over runner scaling
			runGrpcScaling(leaseContext, url, handler)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(leaseTimeout):
		}
	}
}

func runGrpcScaling(ctx context.Context, url url.URL, handler func(ctx context.Context, msg *ftlv1.PullSchemaResponse) error) {
	client := rpc.Dial(ftlv1connect.NewControllerServiceClient, url.String(), log.Error)
	ctx = rpc.ContextWithClient(ctx, client)

	logger := log.FromContext(ctx)
	logger.Debugf("Starting Runner Scaling")
	logger.Debugf("Using FTL endpoint: %s", url.String())

	rpc.RetryStreamingServerStream(ctx, backoff.Backoff{Max: time.Second}, &ftlv1.PullSchemaRequest{}, client.PullSchema, handler, rpc.AlwaysRetry())
	logger.Debugf("Stopped Runner Scaling")
}
