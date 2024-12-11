package lease

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/kong"

	ftllease "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/lease/v1"
	leaseconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/lease/v1/ftlv1connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type Config struct {
	Bind *url.URL `help:"Socket to bind to." default:"http://127.0.0.1:8895" env:"FTL_BIND"`
}

func (c *Config) SetDefaults() {
	if err := kong.ApplyDefaults(c); err != nil {
		panic(err)
	}
}

type service struct {
	lock   sync.Mutex
	leases map[string]*time.Time
}

func Start(ctx context.Context, config Config) error {
	config.SetDefaults()

	logger := log.FromContext(ctx).Scope("lease")
	svc := &service{
		leases: make(map[string]*time.Time),
	}
	ctx = log.ContextWithLogger(ctx, logger)

	logger.Debugf("Lease service listening on: %s", config.Bind)
	err := rpc.Serve(ctx, config.Bind,
		rpc.GRPC(leaseconnect.NewLeaseServiceHandler, svc),
		rpc.HTTP("/", http.NotFoundHandler()),
		rpc.PProf(),
	)
	if err != nil {
		return fmt.Errorf("lease service stopped serving: %w", err)
	}
	return nil
}

func (s *service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *service) AcquireLease(ctx context.Context, stream *connect.BidiStream[ftllease.AcquireLeaseRequest, ftllease.AcquireLeaseResponse]) error {
	logger := log.FromContext(ctx)
	logger.Debugf("AcquireLease called")
	c := &leaseClient{
		leases:  make(map[string]*time.Time),
		service: s,
	}
	defer c.clearLeases()
	for {
		msg, err := stream.Receive()
		if err != nil {
			logger.Errorf(err, "Could not receive lease request")
			return fmt.Errorf("could not receive lease request: %w", err)
		}
		logger.Debugf("Acquiring lease for: %v", msg.Key)
		success := c.handleMessage(msg.Key, msg.Ttl.AsDuration())

		if !success {
			return connect.NewError(connect.CodeResourceExhausted, fmt.Errorf("lease already held"))
		}
		if err = stream.Send(&ftllease.AcquireLeaseResponse{}); err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not send lease response: %w", err))
		}
	}
}

// clearLeases removes leases that were owned by a given connection
// they are only removed if the expiration time matches the one in the map
// this is to handle the case of another connection acquiring the lease after it expires
func (c *leaseClient) clearLeases() {
	s := c.service
	s.lock.Lock()
	defer s.lock.Unlock()
	for key, exp := range c.leases {
		if s.leases[key] != nil && *s.leases[key] == *exp {
			delete(s.leases, key)
		}
	}
}

func (c *leaseClient) handleMessage(keys []string, ttl time.Duration) bool {
	s := c.service
	s.lock.Lock()
	defer s.lock.Unlock()
	key := toKey(keys)
	myExisting := c.leases[key]
	realExisting := s.leases[key]
	if myExisting != nil {
		if myExisting != realExisting {
			// The lease expired, we just fail and don't try to re-acquire it
			// Otherwise it is possible another client acquired and released the lease in the meantime
			// so we should make sure this client knows that the lease was not valid for the whole time
			return false
		}
		if myExisting.After(time.Now()) {
			// We already hold the lease
			exp := time.Now().Add(ttl)
			c.leases[key] = &exp
			s.leases[key] = &exp
			return true
		}
		// The lease expired, we just fail and don't try to re-acquire it
		// Otherwise it is possible another client acquired and released the lease in the meantime
		// Unlikely as the ttl is the same, but still possible
		return false
	}
	if realExisting != nil && realExisting.After(time.Now()) {
		// Someone else holds the lease
		return false
	}
	// grab the lease
	exp := time.Now().Add(ttl)
	c.leases[key] = &exp
	s.leases[key] = &exp
	return true
}

func toKey(key []string) string {
	return strings.Join(key, "/")
}

type leaseClient struct {
	// A local record of all leases held by this client, my diverge if they are not renewed in time
	leases  map[string]*time.Time
	service *service
}
