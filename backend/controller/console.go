package controller

import (
	"context"
	"encoding/json"
	"time"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types"
	"github.com/bufbuild/connect-go"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/TBD54566975/ftl/backend/common/log"
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

	sch := &schema.Schema{
		Modules: slices.Map(deployments, func(d dal.Deployment) *schema.Module {
			return d.Schema
		}),
	}

	var modules []*pbconsole.Module
	for _, deployment := range deployments {
		var verbs []*pbconsole.Verb
		var data []*pbconsole.Data

		for _, decl := range deployment.Schema.Decls {
			switch decl := decl.(type) {
			case *schema.Verb:
				//nolint:forcetypeassert
				v := decl.ToProto().(*pschema.Verb)
				verbSchema := schema.VerbToSchema(v)
				dataRef := schema.DataRef{
					Module: deployment.Module,
					Name:   verbSchema.Request.Name,
				}
				jsonRequestSchema, err := schema.DataToJSONSchema(sch, dataRef)
				if err != nil {
					return nil, errors.WithStack(err)
				}
				jsonData, err := json.MarshalIndent(jsonRequestSchema, "", "  ")
				if err != nil {
					return nil, errors.WithStack(err)
				}
				verbs = append(verbs, &pbconsole.Verb{
					Verb:              v,
					Schema:            verbSchema.String(),
					JsonRequestSchema: string(jsonData),
				})
			case *schema.Data:
				//nolint:forcetypeassert
				d := decl.ToProto().(*pschema.Data)
				data = append(data, &pbconsole.Data{
					Data:   d,
					Schema: schema.DataToSchema(d).String(),
				})
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

func (c *ConsoleService) GetEvents(ctx context.Context, req *connect.Request[pbconsole.EventsQuery]) (*connect.Response[pbconsole.GetEventsResponse], error) {
	query, err := eventsQueryProtoToDAL(req.Msg)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if req.Msg.Limit == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("limit must be > 0"))
	}
	limit := int(req.Msg.Limit)

	// Get 1 more than the requested limit to determine if there are more results.
	limitPlusOne := limit + 1

	results, err := c.dal.QueryEvents(ctx, limitPlusOne, query...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var cursor *int64
	// Return only the requested number of results.
	if len(results) > limit {
		results = results[:limit]
		id := results[len(results)-1].GetID()
		cursor = &id
	}

	response := &pbconsole.GetEventsResponse{
		Events: slices.Map(results, eventDALToProto),
		Cursor: cursor,
	}
	return connect.NewResponse(response), nil
}

func (c *ConsoleService) StreamEvents(ctx context.Context, req *connect.Request[pbconsole.StreamEventsRequest], stream *connect.ServerStream[pbconsole.StreamEventsResponse]) error {
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
	var lastEventTime time.Time
	if req.Msg.AfterTime != nil {
		lastEventTime = req.Msg.AfterTime.AsTime()
	} else {
		lastEventTime = time.Now()
	}

	for {
		thisRequestTime := time.Now()
		events, err := c.dal.QueryEvents(ctx, 1000, append(query, dal.FilterTimeRange(thisRequestTime, lastEventTime))...)
		if err != nil {
			return errors.WithStack(err)
		}

		for index, timelineEvent := range events {
			more := len(events) > index+1
			err := stream.Send(&pbconsole.StreamEventsResponse{
				Event: eventDALToProto(timelineEvent),
				More:  more,
			})
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

func callEventToCall(event *dal.CallEvent) *pbconsole.Call {
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

func logEventToLogEntry(event *dal.LogEvent) *pbconsole.LogEntry {
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

func deploymentEventToDeployment(event *dal.DeploymentEvent) *pbconsole.Deployment {
	var eventType pbconsole.DeploymentEventType
	switch event.Type {
	case dal.DeploymentCreated:
		eventType = pbconsole.DeploymentEventType_DEPLOYMENT_CREATED
	case dal.DeploymentUpdated:
		eventType = pbconsole.DeploymentEventType_DEPLOYMENT_UPDATED
	case dal.DeploymentReplaced:
		eventType = pbconsole.DeploymentEventType_DEPLOYMENT_REPLACED
	default:
		panic(errors.Errorf("unknown deployment event type %v", event.Type))
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

func filterEvents[E dal.Event](events []dal.Event) []E {
	var filtered []E
	for _, event := range events {
		if e, ok := event.(E); ok {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func eventsQueryProtoToDAL(pb *pbconsole.EventsQuery) ([]dal.EventFilter, error) {
	var query []dal.EventFilter

	if pb.Order == pbconsole.EventsQuery_DESC {
		query = append(query, dal.FilterDescending())
	}

	for _, filter := range pb.Filters {
		switch filter := filter.Filter.(type) {
		case *pbconsole.EventsQuery_Filter_Deployments:
			deploymentNames := make([]model.DeploymentName, 0, len(filter.Deployments.Deployments))
			for _, deployment := range filter.Deployments.Deployments {
				deploymentName, err := model.ParseDeploymentName(deployment)
				if err != nil {
					return nil, connect.NewError(connect.CodeInvalidArgument, errors.WithStack(err))
				}
				deploymentNames = append(deploymentNames, deploymentName)
			}
			query = append(query, dal.FilterDeployments(deploymentNames...))

		case *pbconsole.EventsQuery_Filter_Requests:
			requestKeys := make([]model.IngressRequestKey, 0, len(filter.Requests.Requests))
			for _, request := range filter.Requests.Requests {
				requestKey, err := model.ParseIngressRequestKey(request)
				if err != nil {
					return nil, connect.NewError(connect.CodeInvalidArgument, errors.WithStack(err))
				}
				requestKeys = append(requestKeys, requestKey)
			}
			query = append(query, dal.FilterRequests(requestKeys...))

		case *pbconsole.EventsQuery_Filter_EventTypes:
			eventTypes := make([]dal.EventType, 0, len(filter.EventTypes.EventTypes))
			for _, eventType := range filter.EventTypes.EventTypes {
				switch eventType {
				case pbconsole.EventType_EVENT_TYPE_CALL:
					eventTypes = append(eventTypes, dal.EventTypeCall)
				case pbconsole.EventType_EVENT_TYPE_LOG:
					eventTypes = append(eventTypes, dal.EventTypeLog)
				case pbconsole.EventType_EVENT_TYPE_DEPLOYMENT:
					eventTypes = append(eventTypes, dal.EventTypeDeployment)
				default:
					return nil, connect.NewError(connect.CodeInvalidArgument, errors.Errorf("unknown event type %v", eventType))
				}
			}

		case *pbconsole.EventsQuery_Filter_LogLevel:
			level := log.Level(filter.LogLevel.LogLevel)
			if level < log.Trace || level > log.Error {
				return nil, connect.NewError(connect.CodeInvalidArgument, errors.Errorf("unknown log level %v", filter.LogLevel.LogLevel))
			}
			query = append(query, dal.FilterLogLevel(level))

		case *pbconsole.EventsQuery_Filter_Time:
			var newerThan, olderThan time.Time
			if filter.Time.NewerThan != nil {
				newerThan = filter.Time.NewerThan.AsTime()
			}
			if filter.Time.OlderThan != nil {
				olderThan = filter.Time.OlderThan.AsTime()
			}
			query = append(query, dal.FilterTimeRange(newerThan, olderThan))

		case *pbconsole.EventsQuery_Filter_Id:
			var lowerThan, higherThan int64
			if filter.Id.LowerThan != nil {
				lowerThan = *filter.Id.LowerThan
			}
			if filter.Id.HigherThan != nil {
				higherThan = *filter.Id.HigherThan
			}
			query = append(query, dal.FilterIDRange(lowerThan, higherThan))
		case *pbconsole.EventsQuery_Filter_Call:
			var sourceModule types.Option[string]
			if filter.Call.SourceModule != nil {
				sourceModule = types.Some(*filter.Call.SourceModule)
			}
			var destVerb types.Option[string]
			if filter.Call.DestVerb != nil {
				destVerb = types.Some(*filter.Call.DestVerb)
			}
			query = append(query, dal.FilterCall(sourceModule, filter.Call.DestModule, destVerb))

		default:
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.Errorf("unknown filter %T", filter))
		}
	}
	return query, nil
}

func eventDALToProto(event dal.Event) *pbconsole.Event {
	switch event := event.(type) {
	case *dal.CallEvent:
		return &pbconsole.Event{
			TimeStamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbconsole.Event_Call{
				Call: callEventToCall(event),
			},
		}

	case *dal.LogEvent:
		return &pbconsole.Event{
			TimeStamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbconsole.Event_Log{
				Log: logEventToLogEntry(event),
			},
		}

	case *dal.DeploymentEvent:
		return &pbconsole.Event{
			TimeStamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbconsole.Event_Deployment{
				Deployment: deploymentEventToDeployment(event),
			},
		}

	default:
		panic(errors.Errorf("unknown event type %T", event))
	}
}
