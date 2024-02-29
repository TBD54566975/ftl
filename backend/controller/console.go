package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/TBD54566975/ftl/backend/controller/dal"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	pbconsole "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/console"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/console/pbconsoleconnect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/slices"
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
		return nil, err
	}

	sch := &schema.Schema{
		Modules: slices.Map(deployments, func(d dal.Deployment) *schema.Module {
			return d.Schema
		}),
	}
	sch.Modules = append(sch.Modules, schema.Builtins())

	var modules []*pbconsole.Module
	for _, deployment := range deployments {
		var verbs []*pbconsole.Verb
		var data []*pbconsole.Data

		for _, decl := range deployment.Schema.Decls {
			switch decl := decl.(type) {
			case *schema.Verb:
				//nolint:forcetypeassert
				v := decl.ToProto().(*schemapb.Verb)
				verbSchema := schema.VerbFromProto(v) // TODO: include all of the types  that the verb references
				var jsonRequestSchema string
				if requestData, ok := verbSchema.Request.(*schema.DataRef); ok {
					jsonSchema, err := schema.DataToJSONSchema(sch, *requestData)
					if err != nil {
						return nil, err
					}
					jsonData, err := json.MarshalIndent(jsonSchema, "", "  ")
					if err != nil {
						return nil, err
					}
					jsonRequestSchema = string(jsonData)
				}
				verbs = append(verbs, &pbconsole.Verb{
					Verb:              v,
					Schema:            verbSchema.String(),
					JsonRequestSchema: jsonRequestSchema,
				})

			case *schema.Data:
				//nolint:forcetypeassert
				d := decl.ToProto().(*schemapb.Data)
				data = append(data, &pbconsole.Data{
					Data:   d,
					Schema: schema.DataFromProto(d).String(),
				})

			case *schema.Bool, *schema.Bytes, *schema.Database, *schema.Float,
				*schema.Int, *schema.Module, *schema.String, *schema.Time,
				*schema.Unit, *schema.Any, *schema.TypeParameter, *schema.Enum:
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
		return nil, err
	}

	if req.Msg.Limit == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("limit must be > 0"))
	}
	limit := int(req.Msg.Limit)

	// Get 1 more than the requested limit to determine if there are more results.
	limitPlusOne := limit + 1

	results, err := c.dal.QueryEvents(ctx, limitPlusOne, query...)
	if err != nil {
		return nil, err
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

	if req.Msg.Query.Limit == 0 {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("limit must be > 0"))
	}

	query, err := eventsQueryProtoToDAL(req.Msg.Query)
	if err != nil {
		return err
	}

	// Default to last 1 day of events
	var lastEventTime time.Time
	for {
		thisRequestTime := time.Now()
		newQuery := query

		if !lastEventTime.IsZero() {
			newQuery = append(newQuery, dal.FilterTimeRange(thisRequestTime, lastEventTime))
		}

		events, err := c.dal.QueryEvents(ctx, int(req.Msg.Query.Limit), newQuery...)
		if err != nil {
			return err
		}

		err = stream.Send(&pbconsole.StreamEventsResponse{
			Events: slices.Map(events, eventDALToProto),
		})
		if err != nil {
			return err
		}
		lastEventTime = thisRequestTime
		select {
		case <-time.After(updateInterval):
		case <-ctx.Done():
			return nil
		}
	}
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
					return nil, connect.NewError(connect.CodeInvalidArgument, err)
				}
				deploymentNames = append(deploymentNames, deploymentName)
			}
			query = append(query, dal.FilterDeployments(deploymentNames...))

		case *pbconsole.EventsQuery_Filter_Requests:
			requestNames := make([]model.RequestName, 0, len(filter.Requests.Requests))
			for _, request := range filter.Requests.Requests {
				_, requestName, err := model.ParseRequestName(request)
				if err != nil {
					return nil, connect.NewError(connect.CodeInvalidArgument, err)
				}
				requestNames = append(requestNames, requestName)
			}
			query = append(query, dal.FilterRequests(requestNames...))

		case *pbconsole.EventsQuery_Filter_EventTypes:
			eventTypes := make([]dal.EventType, 0, len(filter.EventTypes.EventTypes))
			for _, eventType := range filter.EventTypes.EventTypes {
				switch eventType {
				case pbconsole.EventType_EVENT_TYPE_CALL:
					eventTypes = append(eventTypes, dal.EventTypeCall)
				case pbconsole.EventType_EVENT_TYPE_LOG:
					eventTypes = append(eventTypes, dal.EventTypeLog)
				case pbconsole.EventType_EVENT_TYPE_DEPLOYMENT_CREATED:
					eventTypes = append(eventTypes, dal.EventTypeDeploymentCreated)
				case pbconsole.EventType_EVENT_TYPE_DEPLOYMENT_UPDATED:
					eventTypes = append(eventTypes, dal.EventTypeDeploymentUpdated)
				default:
					return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown event type %v", eventType))
				}
			}
			query = append(query, dal.FilterTypes(eventTypes...))

		case *pbconsole.EventsQuery_Filter_LogLevel:
			level := log.Level(filter.LogLevel.LogLevel)
			if level < log.Trace || level > log.Error {
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown log level %v", filter.LogLevel.LogLevel))
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
			query = append(query, dal.FilterTimeRange(olderThan, newerThan))

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
			var sourceModule optional.Option[string]
			if filter.Call.SourceModule != nil {
				sourceModule = optional.Some(*filter.Call.SourceModule)
			}
			var destVerb optional.Option[string]
			if filter.Call.DestVerb != nil {
				destVerb = optional.Some(*filter.Call.DestVerb)
			}
			query = append(query, dal.FilterCall(sourceModule, filter.Call.DestModule, destVerb))

		default:
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown filter %T", filter))
		}
	}
	return query, nil
}

func eventDALToProto(event dal.Event) *pbconsole.Event {
	switch event := event.(type) {
	case *dal.CallEvent:
		var requestName *string
		if r, ok := event.RequestName.Get(); ok {
			rstr := r.String()
			requestName = &rstr
		}
		var sourceVerbRef *schemapb.VerbRef
		if sourceVerb, ok := event.SourceVerb.Get(); ok {
			sourceVerbRef = sourceVerb.ToProto().(*schemapb.VerbRef) //nolint:forcetypeassert
		}
		return &pbconsole.Event{
			TimeStamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbconsole.Event_Call{
				Call: &pbconsole.CallEvent{
					RequestName:    requestName,
					DeploymentName: event.DeploymentName.String(),
					TimeStamp:      timestamppb.New(event.Time),
					SourceVerbRef:  sourceVerbRef,
					DestinationVerbRef: &schemapb.VerbRef{
						Module: event.DestVerb.Module,
						Name:   event.DestVerb.Name,
					},
					Duration: durationpb.New(event.Duration),
					Request:  string(event.Request),
					Response: string(event.Response),
					Error:    event.Error.Ptr(),
					Stack:    event.Stack.Ptr(),
				},
			},
		}

	case *dal.LogEvent:
		var requestName *string
		if r, ok := event.RequestName.Get(); ok {
			rstr := r.String()
			requestName = &rstr
		}
		return &pbconsole.Event{
			TimeStamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbconsole.Event_Log{
				Log: &pbconsole.LogEvent{
					DeploymentName: event.DeploymentName.String(),
					RequestName:    requestName,
					TimeStamp:      timestamppb.New(event.Time),
					LogLevel:       event.Level,
					Attributes:     event.Attributes,
					Message:        event.Message,
					Error:          event.Error.Ptr(),
					Stack:          event.Stack.Ptr(),
				},
			},
		}

	case *dal.DeploymentCreatedEvent:
		var replaced *string
		if r, ok := event.ReplacedDeployment.Get(); ok {
			rstr := r.String()
			replaced = &rstr
		}
		return &pbconsole.Event{
			TimeStamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbconsole.Event_DeploymentCreated{
				DeploymentCreated: &pbconsole.DeploymentCreatedEvent{
					Name:        event.DeploymentName.String(),
					Language:    event.Language,
					ModuleName:  event.ModuleName,
					MinReplicas: int32(event.MinReplicas),
					Replaced:    replaced,
				},
			},
		}
	case *dal.DeploymentUpdatedEvent:
		return &pbconsole.Event{
			TimeStamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbconsole.Event_DeploymentUpdated{
				DeploymentUpdated: &pbconsole.DeploymentUpdatedEvent{
					Name:            event.DeploymentName.String(),
					MinReplicas:     int32(event.MinReplicas),
					PrevMinReplicas: int32(event.PrevMinReplicas),
				},
			},
		}

	default:
		panic(fmt.Errorf("unknown event type %T", event))
	}
}
