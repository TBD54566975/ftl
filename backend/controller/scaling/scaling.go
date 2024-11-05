package scaling

import (
	"context"
	"errors"
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

type RunnerScaling interface {
	Start(ctx context.Context, endpoint url.URL, leaser leases.Leaser) error

	GetEndpointForDeployment(ctx context.Context, module string, deployment string) (optional.Option[url.URL], error)
}

func BeginGrpcScaling(ctx context.Context, url url.URL, leaser leases.Leaser, handler func(ctx context.Context, msg *ftlv1.PullSchemaResponse) error) {
	leaseTimeout := time.Second * 20
	var leaseDone <-chan struct{}

	for {
		// Grab a lease to take control of runner scaling
		lease, leaseContext, err := leaser.AcquireLease(ctx, leases.SystemKey("ftl-scaling", "runner-creation"), leaseTimeout, optional.None[any]())
		if err == nil {
			leaseDone = leaseContext.Done()
			// If we get it then we take over runner scaling
			runGrpcScaling(leaseContext, url, handler)
		} else if !errors.Is(err, leases.ErrConflict) {
			logger := log.FromContext(ctx)
			logger.Errorf(err, "Failed to acquire lease")
			leaseDone = nil
		}
		select {
		case <-ctx.Done():
			if lease != nil {
				err := lease.Release()
				if err != nil {
					logger := log.FromContext(ctx)
					logger.Errorf(err, "Failed to release lease")
				}
			}
			return
		case <-time.After(leaseTimeout):
		case <-leaseDone:
		}
	}
}

func runGrpcScaling(ctx context.Context, url url.URL, handler func(ctx context.Context, msg *ftlv1.PullSchemaResponse) error) {
	client := rpc.Dial(ftlv1connect.NewControllerServiceClient, url.String(), log.Error)
	ctx = rpc.ContextWithClient(ctx, client)

	logger := log.FromContext(ctx)
	logger.Debugf("Starting Runner Scaling")
	logger.Debugf("Using FTL endpoint: %s", url.String())

	rpc.RetryStreamingServerStream(ctx, "local-scaling", backoff.Backoff{Max: time.Second}, &ftlv1.PullSchemaRequest{}, client.PullSchema, handler, rpc.AlwaysRetry())
	logger.Debugf("Stopped Runner Scaling")
}
