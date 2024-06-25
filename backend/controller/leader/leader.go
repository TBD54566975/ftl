// Package leader provides a way to coordinate a single leader and multiple followers.
//
// Coordinator uses factory functions for leaders and followers, creating each as needed.
// Leader and followers conform to the same protocol, abstracting away the difference to callers.
// Each coordinator has a url to advertise the leader to other coordinators if it generates one.
//
// A leader is created when a lease can be acquired in the database.
// Leaders last as long as the lease can be successfully renewed. Leaders should react to the context being cancelled to know
// when they are no longer leading.
//
// A follower is created with with the url of the leader. Followers last as long as the url for the leader has not changed.
// Followers should react to the context being cancelled to know when they are no longer active.

package leader

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/db/dalerrs"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/alecthomas/types/optional"
)

// LeaderFactory is a function that is called whenever a new leader is acquired.
//
// The context provided is tied to the lease and will be cancelled when the leader is no longer leading.
type LeaderFactory[P any] func(ctx context.Context) (P, error)

// FollowerFactory is a function that is called whenever we follow a new leader.
//
// If the new leader has the same url as the previous leader, the existing follower will be used.
type FollowerFactory[P any] func(ctx context.Context, leaderURL *url.URL) (P, error)

type leader[P any] struct {
	value P
	lease leases.Lease
}

type follower[P any] struct {
	value     P
	deadline  time.Time
	url       *url.URL
	cancelCtx context.CancelFunc
}

// Coordinator assigns a single leader for the rest to follow.
//
// P is the protocol that the leader and followers must implement. Callers of Get() will receive a P,
// abstracting away whether they are interacting with a leader or a follower.
type Coordinator[P any] struct {
	// ctx is passed into the follower factory and is the parent context of leader's lease context
	// it is captured at the time of Coordinator creation as the context when getting may be short lived
	ctx context.Context

	advertise *url.URL
	key       leases.Key
	leaser    leases.Leaser
	leaseTTL  time.Duration

	// leader is active leader value is set
	leaderFactory LeaderFactory[P]
	leader        optional.Option[leader[P]]

	followerFactory FollowerFactory[P]
	follower        optional.Option[*follower[P]]

	// mutex protects leader and follower coordination
	mutex sync.Mutex
}

func NewCoordinator[P any](ctx context.Context,
	advertise *url.URL,
	key leases.Key,
	leaser leases.Leaser,
	leaseTTL time.Duration,
	leaderFactory LeaderFactory[P],
	followerFactory FollowerFactory[P]) *Coordinator[P] {
	coordinator := &Coordinator[P]{
		ctx:             ctx,
		leaser:          leaser,
		leaseTTL:        leaseTTL,
		key:             key,
		advertise:       advertise,
		leaderFactory:   leaderFactory,
		followerFactory: followerFactory,
	}
	// Attempt to coordinate proactively without blocking
	go func() {
		_, _ = coordinator.Get() //nolint:errcheck
	}()
	return coordinator
}

// Get returns either a leader or follower
func (c *Coordinator[P]) Get() (leaderOrFollower P, err error) {
	// Can not have multiple Get() calls in parallel as they may conflict with each other.
	c.mutex.Lock()
	defer c.mutex.Unlock()

	logger := log.FromContext(c.ctx)
	if l, ok := c.leader.Get(); ok {
		// currently leading
		return l.value, nil
	}
	if f, ok := c.follower.Get(); ok && time.Now().Before(f.deadline) {
		// currently following
		return f.value, nil
	}

	lease, leaderCtx, leaseErr := c.leaser.AcquireLease(c.ctx, c.key, c.leaseTTL, optional.Some[any](c.advertise.String()))
	if leaseErr == nil {
		// became leader
		c.retireFollower()
		l, err := c.leaderFactory(leaderCtx)
		if err != nil {
			err := lease.Release()
			if err != nil {
				logger.Warnf("could not release lease after failing to create leader for %s: %s", c.key, err)
			}
			return leaderOrFollower, fmt.Errorf("could not create leader for %s: %w", c.key, err)
		}
		c.leader = optional.Some(leader[P]{
			lease: lease,
			value: l,
		})
		go func() {
			c.watchForLeaderExpiration(leaderCtx)
		}()
		logger.Tracef("new leader for %s: %s", c.key, c.advertise)
		return l, nil
	}
	if !errors.Is(leaseErr, leases.ErrConflict) {
		return leaderOrFollower, fmt.Errorf("could not acquire lease for %s: %w", c.key, leaseErr)
	}
	// lease already held
	return c.createFollower()
}

// watchForLeaderExpiration will remove the leader when the lease's context is cancelled due to failure to heartbeat the lease
func (c *Coordinator[P]) watchForLeaderExpiration(ctx context.Context) {
	<-ctx.Done()

	logger := log.FromContext(c.ctx)
	logger.Warnf("removing leader for %s", c.key)

	c.mutex.Lock()
	c.leader = optional.None[leader[P]]()
	c.mutex.Unlock()
}

func (c *Coordinator[P]) createFollower() (out P, err error) {
	var urlString string
	expiry, err := c.leaser.GetLeaseInfo(c.ctx, c.key, &urlString)
	if err != nil {
		if errors.Is(err, dalerrs.ErrNotFound) {
			return out, fmt.Errorf("could not acquire or find lease for %s", c.key)
		}
		return out, fmt.Errorf("could not get lease for %s: %w", c.key, err)
	}
	if urlString == c.advertise.String() {
		// This prevents endless loops after a lease breaks.
		// If we create a follower pointing locally, the receiver will likely try to then call the leader, which starts the loop again.
		return out, fmt.Errorf("could not follow %s leader at own url: %s", c.key, urlString)
	}
	// check if url matches existing follower's url, just with newer deadline
	if f, ok := c.follower.Get(); ok && f.url.String() == urlString {
		f.deadline = expiry
		return f.value, nil
	}
	c.retireFollower()
	url, err := url.Parse(urlString)
	if err != nil {
		return out, fmt.Errorf("could not parse leader url for %s: %w", c.key, err)
	}
	followerContext, cancel := context.WithCancel(c.ctx)
	f, err := c.followerFactory(followerContext, url)
	if err != nil {
		cancel()
		return out, fmt.Errorf("could not generate follower for %s: %w", c.key, err)
	}
	c.follower = optional.Some(&follower[P]{
		value:     f,
		deadline:  expiry,
		url:       url,
		cancelCtx: cancel,
	})
	return f, nil
}

func (c *Coordinator[P]) retireFollower() {
	f, ok := c.follower.Get()
	if !ok {
		return
	}
	f.cancelCtx()
	c.follower = optional.None[*follower[P]]()
}
