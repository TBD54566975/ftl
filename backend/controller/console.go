package controller

import (
	"context"
	"sort"
	"time"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"

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
			Name:     deployment.Module,
			Language: deployment.Language,
			Verbs:    verbs,
			Data:     data,
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

type ByTimeStamp []*pbconsole.TimelineEntry

func (a ByTimeStamp) Len() int           { return len(a) }
func (a ByTimeStamp) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTimeStamp) Less(i, j int) bool { return a[i].TimeStamp < a[j].TimeStamp }

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
				TimeStamp: deployment.CreatedAt.Unix(),
				Entry: &pbconsole.TimelineEntry_Deployment{
					Deployment: &pbconsole.Deployment{
						Key:         deployment.Key.String(),
						Language:    deployment.Language,
						Name:        deployment.Module,
						MinReplicas: int32(deployment.MinReplicas),
					},
				},
			})
		}
	}

	sort.Sort(ByTimeStamp(timelineEntries))

	return connect.NewResponse(&pbconsole.GetTimelineResponse{
		Entries: timelineEntries,
	}), nil
}

func (c *ConsoleService) StreamLogs(ctx context.Context, req *connect.Request[pbconsole.StreamLogsRequest], stream *connect.ServerStream[pbconsole.StreamLogsResponse]) error {
	// Default to 1 second interval if not specified.
	updateInterval := 1 * time.Second
	if req.Msg.UpdateIntervalMs != 0 {
		updateInterval = time.Duration(req.Msg.UpdateIntervalMs) * time.Millisecond
	}

	var maybeDeploymentKey *model.DeploymentKey
	if req.Msg.DeploymentKey != "" {
		deploymentKey, err := model.ParseDeploymentKey(req.Msg.DeploymentKey)
		if err != nil {
			return errors.WithStack(err)
		}
		maybeDeploymentKey = &deploymentKey
	}

	afterID := int64(0)

	for {
		logs, err := c.dal.GetDeploymentLogs(ctx, req.Msg.AfterTime, afterID, maybeDeploymentKey)
		if err != nil {
			return errors.WithStack(err)
		}

		for index, log := range logs {
			err := stream.Send(&pbconsole.StreamLogsResponse{
				Log: &pbconsole.LogEntry{
					DeploymentKey: log.DeploymentKey.String(),
					RunnerKey:     log.RunnerKey.String(),
					TimeStamp:     log.TimeStamp.Unix(),
					LogLevel:      log.Level,
					Attributes:    log.Attributes,
					Message:       log.Message,
					Error:         log.Error,
				},
				More: len(logs) > index+1,
			})
			if err != nil {
				return errors.WithStack(err)
			}
			afterID = log.ID
		}
		select {
		case <-time.After(updateInterval):
		case <-ctx.Done():
			return nil
		}
	}
}

func convertModuleCalls(calls []dal.CallEntry) []*pbconsole.Call {
	return slices.Map(calls, func(call dal.CallEntry) *pbconsole.Call {
		var errorMessage string
		if call.Error != nil {
			errorMessage = call.Error.Error()
		}
		return &pbconsole.Call{
			Id:            call.ID,
			RunnerKey:     call.RunnerKey.String(),
			RequestKey:    call.RequestKey.String(),
			ControllerKey: call.ControllerKey.String(),
			TimeStamp:     call.Time.Unix(),
			SourceModule:  call.SourceVerb.Module,
			SourceVerb:    call.SourceVerb.Name,
			DestModule:    call.DestVerb.Module,
			DestVerb:      call.DestVerb.Name,
			DurationMs:    call.Duration.Milliseconds(),
			Request:       string(call.Request),
			Response:      string(call.Response),
			Error:         errorMessage,
		}
	})
}
