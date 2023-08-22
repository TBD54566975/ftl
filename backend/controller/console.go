package controller

import (
	"context"
	"time"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/TBD54566975/ftl/backend/common/model"
	"github.com/TBD54566975/ftl/backend/common/slices"
	"github.com/TBD54566975/ftl/backend/controller/internal/dal"
	"github.com/TBD54566975/ftl/backend/schema"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	pbconsole "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/console"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/console/pbconsoleconnect"
	pschema "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

type ConsoleService struct {
	dal *dal.DAL
}

var _ pbconsoleconnect.ConsoleServiceHandler = (*ConsoleService)(nil)

func NewConsoleService(dal *dal.DAL) *ConsoleService {
	return &ConsoleService{
		dal: dal,
	}
}

func (*ConsoleService) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (c *ConsoleService) GetModules(ctx context.Context, req *connect.Request[pbconsole.GetModulesRequest]) (*connect.Response[pbconsole.GetModulesResponse], error) {
	deployments, err := c.dal.GetActiveDeployments(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var modules []*pbconsole.Module
	for _, deployment := range deployments {
		var verbs []*pbconsole.Verb
		var data []*pschema.Data

		for _, decl := range deployment.Schema.Decls {
			switch decl := decl.(type) {
			case *schema.Verb:
				//nolint:forcetypeassert
				verbs = append(verbs, &pbconsole.Verb{
					Verb: decl.ToProto().(*pschema.Verb),
				})
			case *schema.Data:
				//nolint:forcetypeassert
				data = append(data, decl.ToProto().(*pschema.Data))
			}
		}

		modules = append(modules, &pbconsole.Module{
			Name:           deployment.Module,
			DeploymentName: deployment.Name.String(),
			Language:       deployment.Language,
			Verbs:          verbs,
			Data:           data,
		})
	}

	return connect.NewResponse(&pbconsole.GetModulesResponse{
		Modules: modules,
	}), nil
}

func (c *ConsoleService) GetCalls(ctx context.Context, req *connect.Request[pbconsole.GetCallsRequest]) (*connect.Response[pbconsole.GetCallsResponse], error) {
	calls, err := c.dal.GetModuleCalls(ctx, []string{req.Msg.Module})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return connect.NewResponse(&pbconsole.GetCallsResponse{
		Calls: convertModuleCalls(calls[schema.VerbRef{
			Module: req.Msg.Module,
			Name:   req.Msg.Verb,
		}]),
	}), nil
}

func (c *ConsoleService) GetRequestCalls(ctx context.Context, req *connect.Request[pbconsole.GetRequestCallsRequest]) (*connect.Response[pbconsole.GetRequestCallsResponse], error) {
	requestKey, err := model.ParseIngressRequestKey(req.Msg.RequestKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	calls, err := c.dal.GetRequestCalls(ctx, requestKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return connect.NewResponse(&pbconsole.GetRequestCallsResponse{
		Calls: convertModuleCalls(calls),
	}), nil
}

func (c *ConsoleService) StreamTimeline(ctx context.Context, req *connect.Request[pbconsole.StreamTimelineRequest], stream *connect.ServerStream[pbconsole.StreamTimelineResponse]) error {
	// Default to 1 second interval if not specified.
	updateInterval := 1 * time.Second
	if req.Msg.UpdateInterval != nil && req.Msg.UpdateInterval.AsDuration() > time.Second { // Minimum 1s interval.
		updateInterval = req.Msg.UpdateInterval.AsDuration()
	}

	var query []dal.EventFilter
	if req.Msg.DeploymentName != "" {
		deploymentName, err := model.ParseDeploymentName(req.Msg.DeploymentName)
		if err != nil {
			return errors.WithStack(err)
		}
		query = append(query, dal.FilterDeployments(deploymentName))
	}

	lastEventTime := req.Msg.AfterTime.AsTime()
	for {
		thisRequestTime := time.Now()
		events, err := c.dal.QueryEvents(ctx, lastEventTime, thisRequestTime, query...)
		if err != nil {
			return errors.WithStack(err)
		}

		timelineEvents := filterTimelineEvents(events)
		for index, timelineEvent := range timelineEvents {
			more := len(events) > index+1
			var err error

			switch event := timelineEvent.(type) {
			case *dal.CallEvent:
				err = stream.Send(&pbconsole.StreamTimelineResponse{
					TimeStamp: timestamppb.New(event.Time),
					Entry: &pbconsole.StreamTimelineResponse_Call{
						Call: callEventToCall(*event),
					},
					More: more,
				})
			case *dal.LogEvent:
				err = stream.Send(&pbconsole.StreamTimelineResponse{
					TimeStamp: timestamppb.New(event.Time),
					Entry: &pbconsole.StreamTimelineResponse_Log{
						Log: logEventToLogEntry(*event),
					},
					More: more,
				})
			case *dal.DeploymentEvent:
				err = stream.Send(&pbconsole.StreamTimelineResponse{
					TimeStamp: timestamppb.New(event.Time),
					Entry: &pbconsole.StreamTimelineResponse_Deployment{
						Deployment: deploymentEventToDeployment(*event),
					},
					More: more,
				})
			}

			if err != nil {
				return errors.WithStack(err)
			}
		}
		lastEventTime = thisRequestTime
		select {
		case <-time.After(updateInterval):
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *ConsoleService) StreamLogs(ctx context.Context, req *connect.Request[pbconsole.StreamLogsRequest], stream *connect.ServerStream[pbconsole.StreamLogsResponse]) error {
	// Default to 1 second interval if not specified.
	updateInterval := 1 * time.Second
	if req.Msg.UpdateInterval != nil && req.Msg.UpdateInterval.AsDuration() > time.Second { // Minimum 1s interval.
		updateInterval = req.Msg.UpdateInterval.AsDuration()
	}

	var query []dal.EventFilter
	if req.Msg.DeploymentName != "" {
		deploymentName, err := model.ParseDeploymentName(req.Msg.DeploymentName)
		if err != nil {
			return errors.WithStack(err)
		}
		query = append(query, dal.FilterDeployments(deploymentName))
	}

	lastLogTime := req.Msg.AfterTime.AsTime()
	for {
		thisRequestTime := time.Now()
		events, err := c.dal.QueryEvents(ctx, lastLogTime, thisRequestTime, query...)
		if err != nil {
			return errors.WithStack(err)
		}

		logEvents := filterLogEvents(events)
		for index, log := range logEvents {
			var requestKey *string
			if r, ok := log.RequestKey.Get(); ok {
				rstr := r.String()
				requestKey = &rstr
			}

			err := stream.Send(&pbconsole.StreamLogsResponse{
				Log: &pbconsole.LogEntry{
					DeploymentName: log.DeploymentName.String(),
					RequestKey:     requestKey,
					TimeStamp:      timestamppb.New(log.Time),
					LogLevel:       log.Level,
					Attributes:     log.Attributes,
					Message:        log.Message,
					Error:          log.Error.Ptr(),
				},
				More: len(logEvents) > index+1,
			})
			if err != nil {
				return errors.WithStack(err)
			}
		}
		lastLogTime = thisRequestTime
		select {
		case <-time.After(updateInterval):
		case <-ctx.Done():
			return nil
		}
	}
}

func convertModuleCalls(calls []dal.CallEvent) []*pbconsole.Call {
	return slices.Map(calls, callEventToCall)
}

func callEventToCall(event dal.CallEvent) *pbconsole.Call {
	var requestKey *string
	if r, ok := event.RequestKey.Get(); ok {
		rstr := r.String()
		requestKey = &rstr
	}
	var sourceVerbRef *pschema.VerbRef
	if sourceVerb, ok := event.SourceVerb.Get(); ok {
		sourceVerbRef = sourceVerb.ToProto().(*pschema.VerbRef) //nolint:forcetypeassert
	}
	return &pbconsole.Call{
		RequestKey:     requestKey,
		DeploymentName: event.DeploymentName.String(),
		TimeStamp:      timestamppb.New(event.Time),
		SourceVerbRef:  sourceVerbRef,
		DestinationVerbRef: &pschema.VerbRef{
			Module: event.DestVerb.Module,
			Name:   event.DestVerb.Name,
		},
		Duration: durationpb.New(event.Duration),
		Request:  string(event.Request),
		Response: string(event.Response),
		Error:    event.Error.Ptr(),
	}
}

func logEventToLogEntry(event dal.LogEvent) *pbconsole.LogEntry {
	var requestKey *string
	if r, ok := event.RequestKey.Get(); ok {
		rstr := r.String()
		requestKey = &rstr
	}
	return &pbconsole.LogEntry{
		DeploymentName: event.DeploymentName.String(),
		RequestKey:     requestKey,
		TimeStamp:      timestamppb.New(event.Time),
		LogLevel:       event.Level,
		Attributes:     event.Attributes,
		Message:        event.Message,
		Error:          event.Error.Ptr(),
	}
}

func deploymentEventToDeployment(event dal.DeploymentEvent) *pbconsole.Deployment {
	var eventType pbconsole.DeploymentEventType
	switch event.Type {
	case dal.DeploymentCreated:
		eventType = pbconsole.DeploymentEventType_DEPLOYMENT_CREATED
	case dal.DeploymentUpdated:
		eventType = pbconsole.DeploymentEventType_DEPLOYMENT_UPDATED
	case dal.DeploymentReplaced:
		eventType = pbconsole.DeploymentEventType_DEPLOYMENT_REPLACED
	}

	var replaced *string
	if r, ok := event.ReplacedDeployment.Get(); ok {
		rstr := r.String()
		replaced = &rstr
	}
	return &pbconsole.Deployment{
		Name:        event.DeploymentName.String(),
		Language:    event.Language,
		ModuleName:  event.ModuleName,
		MinReplicas: 0,
		EventType:   eventType,
		Replaced:    replaced,
	}
}

func filterLogEvents(events []dal.Event) []*dal.LogEvent {
	var filtered []*dal.LogEvent
	for _, event := range events {
		if log, ok := event.(*dal.LogEvent); ok {
			filtered = append(filtered, log)
		}
	}
	return filtered
}

func filterTimelineEvents(events []dal.Event) []dal.Event {
	var filtered []dal.Event
	for _, event := range events {
		if _, ok := event.(*dal.LogEvent); ok {
			filtered = append(filtered, event)
		}
		if _, ok := event.(*dal.CallEvent); ok {
			filtered = append(filtered, event)
		}
		if _, ok := event.(*dal.DeploymentEvent); ok {
			filtered = append(filtered, event)
		}
	}
	return filtered
}
