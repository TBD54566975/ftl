package timeline

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/kong"
	"google.golang.org/protobuf/types/known/timestamppb"

	timelinepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	timelineconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1/timelinev1connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema/schemaeventsource"
	"github.com/TBD54566975/ftl/internal/slices"
)

type Config struct {
	Bind *url.URL `help:"Socket to bind to." default:"http://127.0.0.1:8894" env:"FTL_BIND"`
}

func (c *Config) SetDefaults() {
	if err := kong.ApplyDefaults(c); err != nil {
		panic(err)
	}
}

type service struct {
	lock   sync.RWMutex
	nextID int
	events []*timelinepb.Event
}

func Start(ctx context.Context, config Config, schemaEventSource schemaeventsource.EventSource) error {
	config.SetDefaults()

	logger := log.FromContext(ctx).Scope("timeline")
	svc := &service{}

	logger.Debugf("Timeline service listening on: %s", config.Bind)
	err := rpc.Serve(ctx, config.Bind,
		rpc.GRPC(timelineconnect.NewTimelineServiceHandler, svc),
	)
	if err != nil {
		return fmt.Errorf("timeline service stopped serving: %w", err)
	}
	return nil
}

func (s *service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *service) CreateEvent(ctx context.Context, req *connect.Request[timelinepb.CreateEventRequest]) (*connect.Response[timelinepb.CreateEventResponse], error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	event := &timelinepb.Event{
		Id:        int64(s.nextID),
		TimeStamp: timestamppb.Now(),
	}
	switch entry := req.Msg.Entry.(type) {
	case *timelinepb.CreateEventRequest_Log:
		event.Entry = &timelinepb.Event_Log{
			Log: entry.Log,
		}
	case *timelinepb.CreateEventRequest_Call:
		event.Entry = &timelinepb.Event_Call{
			Call: entry.Call,
		}
	case *timelinepb.CreateEventRequest_DeploymentCreated:
		event.Entry = &timelinepb.Event_DeploymentCreated{
			DeploymentCreated: entry.DeploymentCreated,
		}
	case *timelinepb.CreateEventRequest_DeploymentUpdated:
		event.Entry = &timelinepb.Event_DeploymentUpdated{
			DeploymentUpdated: entry.DeploymentUpdated,
		}
	case *timelinepb.CreateEventRequest_Ingress:
		event.Entry = &timelinepb.Event_Ingress{
			Ingress: entry.Ingress,
		}
	case *timelinepb.CreateEventRequest_CronScheduled:
		event.Entry = &timelinepb.Event_CronScheduled{
			CronScheduled: entry.CronScheduled,
		}
	case *timelinepb.CreateEventRequest_AsyncExecute:
		event.Entry = &timelinepb.Event_AsyncExecute{
			AsyncExecute: entry.AsyncExecute,
		}
	case *timelinepb.CreateEventRequest_PubsubPublish:
		event.Entry = &timelinepb.Event_PubsubPublish{
			PubsubPublish: entry.PubsubPublish,
		}
	case *timelinepb.CreateEventRequest_PubsubConsume:
		event.Entry = &timelinepb.Event_PubsubConsume{
			PubsubConsume: entry.PubsubConsume,
		}
	}
	s.events = append(s.events, event)
	s.nextID++
	return connect.NewResponse(&timelinepb.CreateEventResponse{}), nil
}

func (s *service) GetTimeline(ctx context.Context, req *connect.Request[timelinepb.GetTimelineRequest]) (*connect.Response[timelinepb.GetTimelineResponse], error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	filters, limit, ascending, err := filtersFromRequest(req.Msg)
	if err != nil {
		return nil, err
	}

	results := []*timelinepb.Event{}

	var firstIdx, step int
	var idxCheck func(int) bool
	if ascending {
		firstIdx = 0
		step = 1
		idxCheck = func(i int) bool { return i < len(s.events) }
	} else {
		firstIdx = len(s.events) - 1
		step = -1
		idxCheck = func(i int) bool { return i >= 0 }
	}
	for i := firstIdx; idxCheck(i); i += step {
		event := s.events[i]
		_, didNotMatchAFilter := slices.Find(filters, func(filter TimelineFilter) bool {
			return !filter(event)
		})
		for didNotMatchAFilter {
			continue
		}
		results = append(results, s.events[i])
		if limit, ok := limit.Get(); ok && limit != 0 && len(results) >= limit {
			break
		}
	}
	return connect.NewResponse(&timelinepb.GetTimelineResponse{
		Events: results,
	}), nil
}

func (s *service) DeleteOldEvents(ctx context.Context, req *connect.Request[timelinepb.DeleteOldEventsRequest]) (*connect.Response[timelinepb.DeleteOldEventsResponse], error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	// Events that match all these filters will be deleted
	cutoff := time.Now().Add(-1 * time.Duration(req.Msg.AgeSeconds) * time.Second)
	deletionFilters := []TimelineFilter{
		FilterTypes(&timelinepb.GetTimelineRequest_EventTypeFilter{
			EventTypes: []timelinepb.EventType{req.Msg.EventType},
		}),
		FilterTimeRange(&timelinepb.GetTimelineRequest_TimeFilter{
			OlderThan: timestamppb.New(cutoff),
		}),
	}

	filtered := []*timelinepb.Event{}
	for _, event := range s.events {
		_, didNotMatchAFilter := slices.Find(deletionFilters, func(filter TimelineFilter) bool {
			return !filter(event)
		})
		if didNotMatchAFilter {
			filtered = append(filtered, event)
		}
	}
	s.events = filtered
	return connect.NewResponse(&timelinepb.DeleteOldEventsResponse{}), nil
}
