package controller

import (
	"context"
	"sort"
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

func (c *ConsoleService) GetTimeline(ctx context.Context, req *connect.Request[pbconsole.GetTimelineRequest]) (*connect.Response[pbconsole.GetTimelineResponse], error) {
	var timelineEntries []*pbconsole.TimelineEntry

	dbCalls, err := c.dal.GetModuleCalls(ctx, []string{req.Msg.Module})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for _, callEntries := range dbCalls {
		calls := convertModuleCalls(callEntries)
		for _, call := range calls {
			timelineEntries = append(timelineEntries, &pbconsole.TimelineEntry{
				TimeStamp: call.TimeStamp,
				Entry: &pbconsole.TimelineEntry_Call{
					Call: call,
				},
			})
		}
	}

	deployments, err := c.dal.GetActiveDeployments(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for _, deployment := range deployments {
		if deployment.Module == req.Msg.Module {
			timelineEntries = append(timelineEntries, &pbconsole.TimelineEntry{
				TimeStamp: timestamppb.New(deployment.CreatedAt),
				Entry: &pbconsole.TimelineEntry_Deployment{
					Deployment: &pbconsole.Deployment{
						Name:        deployment.Name.String(),
						Language:    deployment.Language,
						ModuleName:  deployment.Module,
						MinReplicas: int32(deployment.MinReplicas),
					},
				},
			})
		}
	}

	sort.Slice(timelineEntries, func(i, j int) bool {
		return timelineEntries[i].TimeStamp.AsTime().Before(timelineEntries[j].TimeStamp.AsTime())
	})

	return connect.NewResponse(&pbconsole.GetTimelineResponse{
		Entries: timelineEntries,
	}), nil
}

func (c *ConsoleService) StreamLogs(ctx context.Context, req *connect.Request[pbconsole.StreamLogsRequest], stream *connect.ServerStream[pbconsole.StreamLogsResponse]) error {
	// Default to 1 second interval if not specified.
	updateInterval := 1 * time.Second
	if req.Msg.UpdateInterval != nil && req.Msg.UpdateInterval.AsDuration() > time.Second { // Minimum 1s interval.
		updateInterval = req.Msg.UpdateInterval.AsDuration()
	}

	query := []dal.EventFilter{}
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
	return slices.Map(calls, func(call dal.CallEvent) *pbconsole.Call {
		var errorMessage string
		if call.Error != nil {
			errorMessage = call.Error.Error()
		}
		return &pbconsole.Call{
			RequestKey:     call.RequestKey.String(),
			DeploymentName: call.DeploymentName.String(),
			TimeStamp:      timestamppb.New(call.Time),
			SourceModule:   call.SourceVerb.Module,
			SourceVerb:     call.SourceVerb.Name,
			DestModule:     call.DestVerb.Module,
			DestVerb:       call.DestVerb.Name,
			Duration:       durationpb.New(call.Duration),
			Request:        string(call.Request),
			Response:       string(call.Response),
			Error:          errorMessage,
		}
	})
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
