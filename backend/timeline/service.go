package timeline

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/types/known/timestamppb"

	timelinepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	timelineconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1/timelinev1connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
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

func Start(ctx context.Context, config Config) error {
	config.SetDefaults()

	logger := log.FromContext(ctx).Scope("timeline")
	svc := &service{
		events: make([]*timelinepb.Event, 0),
		nextID: 0,
	}

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

func (s *service) CreateEvents(ctx context.Context, req *connect.Request[timelinepb.CreateEventsRequest]) (*connect.Response[timelinepb.CreateEventsResponse], error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	entries := make([]*timelinepb.CreateEventsRequest_EventEntry, 0, len(req.Msg.Entries))
	entries = append(entries, req.Msg.Entries...)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.AsTime().Before(entries[j].Timestamp.AsTime())
	})

	for _, entry := range entries {
		if entry.Timestamp == nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("timestamp is required"))
		}
		event := &timelinepb.Event{
			Id:        int64(s.nextID),
			Timestamp: entry.Timestamp,
		}
		switch entry := entry.Entry.(type) {
		case *timelinepb.CreateEventsRequest_EventEntry_Log:
			event.Entry = &timelinepb.Event_Log{
				Log: entry.Log,
			}
		case *timelinepb.CreateEventsRequest_EventEntry_Call:
			event.Entry = &timelinepb.Event_Call{
				Call: entry.Call,
			}
		case *timelinepb.CreateEventsRequest_EventEntry_DeploymentCreated:
			event.Entry = &timelinepb.Event_DeploymentCreated{
				DeploymentCreated: entry.DeploymentCreated,
			}
		case *timelinepb.CreateEventsRequest_EventEntry_DeploymentUpdated:
			event.Entry = &timelinepb.Event_DeploymentUpdated{
				DeploymentUpdated: entry.DeploymentUpdated,
			}
		case *timelinepb.CreateEventsRequest_EventEntry_Ingress:
			event.Entry = &timelinepb.Event_Ingress{
				Ingress: entry.Ingress,
			}
		case *timelinepb.CreateEventsRequest_EventEntry_CronScheduled:
			event.Entry = &timelinepb.Event_CronScheduled{
				CronScheduled: entry.CronScheduled,
			}
		case *timelinepb.CreateEventsRequest_EventEntry_AsyncExecute:
			event.Entry = &timelinepb.Event_AsyncExecute{
				AsyncExecute: entry.AsyncExecute,
			}
		case *timelinepb.CreateEventsRequest_EventEntry_PubsubPublish:
			event.Entry = &timelinepb.Event_PubsubPublish{
				PubsubPublish: entry.PubsubPublish,
			}
		case *timelinepb.CreateEventsRequest_EventEntry_PubsubConsume:
			event.Entry = &timelinepb.Event_PubsubConsume{
				PubsubConsume: entry.PubsubConsume,
			}
		}
		s.events = append(s.events, event)
		s.nextID++
	}
	return connect.NewResponse(&timelinepb.CreateEventsResponse{}), nil
}

func (s *service) GetTimeline(ctx context.Context, req *connect.Request[timelinepb.GetTimelineRequest]) (*connect.Response[timelinepb.GetTimelineResponse], error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	filters, ascending := filtersFromRequest(req.Msg)
	if req.Msg.Limit == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("limit must be > 0"))
	}
	// Get 1 more than the requested limit to determine if there are more results.
	limit := int(req.Msg.Limit)
	fetchLimit := limit + 1

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
		if didNotMatchAFilter {
			continue
		}
		results = append(results, s.events[i])
		if fetchLimit != 0 && len(results) >= fetchLimit {
			break
		}
	}

	var cursor *int64
	// Return only the requested number of results.
	if len(results) > limit {
		id := results[len(results)-1].Id
		results = results[:limit]
		cursor = &id
	}
	return connect.NewResponse(&timelinepb.GetTimelineResponse{
		Events: results,
		Cursor: cursor,
	}), nil
}

func (s *service) StreamTimeline(ctx context.Context, req *connect.Request[timelinepb.StreamTimelineRequest], stream *connect.ServerStream[timelinepb.StreamTimelineResponse]) error {
	// Default to 1 second interval if not specified.
	updateInterval := 1 * time.Second
	if req.Msg.UpdateInterval != nil && req.Msg.UpdateInterval.AsDuration() > time.Second { // Minimum 1s interval.
		updateInterval = req.Msg.UpdateInterval.AsDuration()
	}

	if req.Msg.Query.Limit == 0 {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("limit must be > 0"))
	}

	_, ascending := filtersFromRequest(req.Msg.Query)

	timelineReq := req.Msg.Query
	// Default to last 1 day of events
	var lastEventId optional.Option[int64]
	for {
		newQuery := timelineReq
		// We always want ascending order for the underlying query.
		newQuery.Order = timelinepb.GetTimelineRequest_ORDER_ASC
		if _, ok := lastEventId.Get(); ok {
			newQuery.Filters = append(newQuery.Filters, &timelinepb.GetTimelineRequest_Filter{
				Filter: &timelinepb.GetTimelineRequest_Filter_Id{
					Id: &timelinepb.GetTimelineRequest_IDFilter{
						HigherThan: lastEventId.Ptr(),
					},
				},
			})
		}

		resp, err := s.GetTimeline(ctx, connect.NewRequest(newQuery))
		if err != nil {
			return fmt.Errorf("failed to get timeline: %w", err)
		}

		newEvents := make([]*timelinepb.Event, 0, len(resp.Msg.Events))
		for _, event := range resp.Msg.Events {
			if lastEventId, ok := lastEventId.Get(); !ok || event.Id != lastEventId {
				// This is not a duplicate event.
				newEvents = append(newEvents, event)
			}
		}
		if len(newEvents) > 0 {
			lastEventId = optional.Some(newEvents[len(newEvents)-1].Id)

			if !ascending {
				// Original query was for descending order, so reverse the events.
				slices.Reverse(newEvents)
			}
			err = stream.Send(&timelinepb.StreamTimelineResponse{
				Events: newEvents,
			})
			if err != nil {
				return fmt.Errorf("failed to get timeline events: %w", err)
			}

		}
		select {
		case <-time.After(updateInterval):
		case <-ctx.Done():
			return nil
		}
	}
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
	deleted := int64(0)
	for _, event := range s.events {
		_, didNotMatchAFilter := slices.Find(deletionFilters, func(filter TimelineFilter) bool {
			return !filter(event)
		})
		if didNotMatchAFilter {
			filtered = append(filtered, event)
		} else {
			deleted++
		}
	}
	s.events = filtered
	return connect.NewResponse(&timelinepb.DeleteOldEventsResponse{
		DeletedCount: deleted,
	}), nil
}
