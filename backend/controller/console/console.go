package console

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
	dalmodel "github.com/TBD54566975/ftl/backend/controller/dal/model"
	"github.com/TBD54566975/ftl/backend/controller/timeline"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	pbconsole "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/console"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/console/pbconsoleconnect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/buildengine"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/slices"
)

type ConsoleService struct {
	dal      *dal.DAL
	timeline *timeline.Service
}

var _ pbconsoleconnect.ConsoleServiceHandler = (*ConsoleService)(nil)

func NewService(dal *dal.DAL, timeline *timeline.Service) *ConsoleService {
	return &ConsoleService{
		dal:      dal,
		timeline: timeline,
	}
}

func (*ConsoleService) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func visitNode(sch *schema.Schema, n schema.Node, verbString *string) error {
	return schema.Visit(n, func(n schema.Node, next func() error) error {
		switch n := n.(type) {
		case *schema.Ref:
			if decl, ok := sch.Resolve(n).Get(); ok {
				*verbString += decl.String() + "\n\n"
				err := visitNode(sch, decl, verbString)
				if err != nil {
					return err
				}
			}

		default:
		}
		return next()
	})
}

func verbSchemaString(sch *schema.Schema, verb *schema.Verb) (string, error) {
	var verbString string
	err := visitNode(sch, verb.Request, &verbString)
	if err != nil {
		return "", err
	}
	// Don't print the response if it's the same as the request.
	if !verb.Response.Equal(verb.Request) {
		err = visitNode(sch, verb.Response, &verbString)
		if err != nil {
			return "", err
		}
	}
	verbString += verb.String()
	return verbString, nil
}

func (c *ConsoleService) GetModules(ctx context.Context, req *connect.Request[pbconsole.GetModulesRequest]) (*connect.Response[pbconsole.GetModulesResponse], error) {
	deployments, err := c.dal.GetDeploymentsWithMinReplicas(ctx)
	if err != nil {
		return nil, err
	}

	sch := &schema.Schema{
		Modules: slices.Map(deployments, func(d dalmodel.Deployment) *schema.Module {
			return d.Schema
		}),
	}
	sch.Modules = append(sch.Modules, schema.Builtins())

	var modules []*pbconsole.Module
	for _, deployment := range deployments {
		var verbs []*pbconsole.Verb
		var data []*pbconsole.Data
		var secrets []*pbconsole.Secret
		var configs []*pbconsole.Config

		for _, decl := range deployment.Schema.Decls {
			switch decl := decl.(type) {
			case *schema.Verb:
				//nolint:forcetypeassert
				v := decl.ToProto().(*schemapb.Verb)
				verbSchema := schema.VerbFromProto(v)
				var jsonRequestSchema string
				if verbSchema.Request != nil {
					if requestData, ok := verbSchema.Request.(*schema.Ref); ok {
						jsonSchema, err := schema.RequestResponseToJSONSchema(sch, *requestData)
						if err != nil {
							return nil, err
						}
						jsonData, err := json.MarshalIndent(jsonSchema, "", "  ")
						if err != nil {
							return nil, err
						}
						jsonRequestSchema = string(jsonData)
					}
				}

				schemaString, err := verbSchemaString(sch, decl)
				if err != nil {
					return nil, err
				}
				verbs = append(verbs, &pbconsole.Verb{
					Verb:              v,
					Schema:            schemaString,
					JsonRequestSchema: jsonRequestSchema,
				})

			case *schema.Data:
				//nolint:forcetypeassert
				d := decl.ToProto().(*schemapb.Data)
				data = append(data, &pbconsole.Data{
					Data:   d,
					Schema: schema.DataFromProto(d).String(),
				})

			case *schema.Secret:
				//nolint:forcetypeassert
				s := decl.ToProto().(*schemapb.Secret)
				secrets = append(secrets, &pbconsole.Secret{
					Secret: s,
				})

			case *schema.Config:
				//nolint:forcetypeassert
				c := decl.ToProto().(*schemapb.Config)
				configs = append(configs, &pbconsole.Config{
					Config: c,
				})

			case *schema.Database, *schema.Enum, *schema.TypeAlias, *schema.FSM, *schema.Topic, *schema.Subscription:
			}
		}

		modules = append(modules, &pbconsole.Module{
			Name:          deployment.Module,
			DeploymentKey: deployment.Key.String(),
			Language:      deployment.Language,
			Verbs:         verbs,
			Data:          data,
			Secrets:       secrets,
			Configs:       configs,
			Schema:        deployment.Schema.String(),
		})
	}

	sorted, err := buildengine.TopologicalSort(graph(sch))
	if err != nil {
		return nil, fmt.Errorf("failed to sort modules: %w", err)
	}
	topology := &pbconsole.Topology{
		Levels: make([]*pbconsole.TopologyGroup, len(sorted)),
	}
	for i, level := range sorted {
		group := &pbconsole.TopologyGroup{
			Modules: level,
		}
		topology.Levels[i] = group
	}

	return connect.NewResponse(&pbconsole.GetModulesResponse{
		Modules:  modules,
		Topology: topology,
	}), nil
}

func (c *ConsoleService) GetConfig(ctx context.Context, req *connect.Request[pbconsole.GetDeclRequest]) (*connect.Response[pbconsole.Config], error) {
	decl, err := c.findDecl(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	//nolint:forcetypeassert
	config, ok := decl.ToProto().(*schemapb.Config)
	if !ok {
		return nil, fmt.Errorf("could not convert config %s.%s to proto", req.Msg.ModuleName, req.Msg.DeclName)
	}
	return connect.NewResponse(&pbconsole.Config{
		Config: config,
	}), nil
}

func (c *ConsoleService) GetData(ctx context.Context, req *connect.Request[pbconsole.GetDeclRequest]) (*connect.Response[pbconsole.Data], error) {
	decl, err := c.findDecl(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	//nolint:forcetypeassert
	data, ok := decl.ToProto().(*schemapb.Data)
	if !ok {
		return nil, fmt.Errorf("could not convert data %s.%s to proto", req.Msg.ModuleName, req.Msg.DeclName)
	}
	return connect.NewResponse(&pbconsole.Data{
		Data:   data,
		Schema: schema.DataFromProto(data).String(),
	}), nil
}

func (c *ConsoleService) GetDatabase(ctx context.Context, req *connect.Request[pbconsole.GetDeclRequest]) (*connect.Response[pbconsole.Database], error) {
	decl, err := c.findDecl(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	//nolint:forcetypeassert
	database, ok := decl.ToProto().(*schemapb.Database)
	if !ok {
		return nil, fmt.Errorf("could not convert database %s.%s to proto", req.Msg.ModuleName, req.Msg.DeclName)
	}
	return connect.NewResponse(&pbconsole.Database{
		Database: database,
	}), nil
}

func (c *ConsoleService) GetEnum(ctx context.Context, req *connect.Request[pbconsole.GetDeclRequest]) (*connect.Response[pbconsole.Enum], error) {
	decl, err := c.findDecl(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	//nolint:forcetypeassert
	enum, ok := decl.ToProto().(*schemapb.Enum)
	if !ok {
		return nil, fmt.Errorf("could not convert enum %s.%s to proto", req.Msg.ModuleName, req.Msg.DeclName)
	}
	return connect.NewResponse(&pbconsole.Enum{
		Enum: enum,
	}), nil
}

func (c *ConsoleService) GetFSM(ctx context.Context, req *connect.Request[pbconsole.GetDeclRequest]) (*connect.Response[pbconsole.FSM], error) {
	decl, err := c.findDecl(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	//nolint:forcetypeassert
	fsm, ok := decl.ToProto().(*schemapb.FSM)
	if !ok {
		return nil, fmt.Errorf("could not convert FSM %s.%s to proto", req.Msg.ModuleName, req.Msg.DeclName)
	}
	return connect.NewResponse(&pbconsole.FSM{
		Fsm: fsm,
	}), nil
}

func (c *ConsoleService) GetTopic(ctx context.Context, req *connect.Request[pbconsole.GetDeclRequest]) (*connect.Response[pbconsole.Topic], error) {
	decl, err := c.findDecl(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	//nolint:forcetypeassert
	topic, ok := decl.ToProto().(*schemapb.Topic)
	if !ok {
		return nil, fmt.Errorf("could not convert topic %s.%s to proto", req.Msg.ModuleName, req.Msg.DeclName)
	}
	return connect.NewResponse(&pbconsole.Topic{
		Topic: topic,
	}), nil
}

func (c *ConsoleService) GetTypeAlias(ctx context.Context, req *connect.Request[pbconsole.GetDeclRequest]) (*connect.Response[pbconsole.TypeAlias], error) {
	decl, err := c.findDecl(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	//nolint:forcetypeassert
	ta, ok := decl.ToProto().(*schemapb.TypeAlias)
	if !ok {
		return nil, fmt.Errorf("could not convert type alias %s.%s to proto", req.Msg.ModuleName, req.Msg.DeclName)
	}
	return connect.NewResponse(&pbconsole.TypeAlias{
		Typealias: ta,
	}), nil
}

func (c *ConsoleService) GetSecret(ctx context.Context, req *connect.Request[pbconsole.GetDeclRequest]) (*connect.Response[pbconsole.Secret], error) {
	decl, err := c.findDecl(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	//nolint:forcetypeassert
	secret, ok := decl.ToProto().(*schemapb.Secret)
	if !ok {
		return nil, fmt.Errorf("could not convert secret %s.%s to proto", req.Msg.ModuleName, req.Msg.DeclName)
	}
	return connect.NewResponse(&pbconsole.Secret{
		Secret: secret,
	}), nil
}

func (c *ConsoleService) GetSubscription(ctx context.Context, req *connect.Request[pbconsole.GetDeclRequest]) (*connect.Response[pbconsole.Subscription], error) {
	decl, err := c.findDecl(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	//nolint:forcetypeassert
	sub, ok := decl.ToProto().(*schemapb.Subscription)
	if !ok {
		return nil, fmt.Errorf("could not convert subscription %s.%s to proto", req.Msg.ModuleName, req.Msg.DeclName)
	}
	return connect.NewResponse(&pbconsole.Subscription{
		Subscription: sub,
	}), nil
}

func (c *ConsoleService) GetVerb(ctx context.Context, req *connect.Request[pbconsole.GetDeclRequest]) (*connect.Response[pbconsole.Verb], error) {
	deployments, err := c.dal.GetDeploymentsWithMinReplicas(ctx)
	if err != nil {
		return nil, err
	}

	sch := &schema.Schema{
		Modules: slices.Map(deployments, func(d dalmodel.Deployment) *schema.Module {
			return d.Schema
		}),
	}
	sch.Modules = append(sch.Modules, schema.Builtins())

	decl, err := c.findDecl(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	//nolint:forcetypeassert
	v := decl.ToProto().(*schemapb.Verb)
	verbSchema := schema.VerbFromProto(v)
	var jsonRequestSchema string
	if verbSchema.Request != nil {
		if requestData, ok := verbSchema.Request.(*schema.Ref); ok {
			jsonSchema, err := schema.RequestResponseToJSONSchema(sch, *requestData)
			if err != nil {
				return nil, err
			}
			jsonData, err := json.MarshalIndent(jsonSchema, "", "  ")
			if err != nil {
				return nil, err
			}
			jsonRequestSchema = string(jsonData)
		}
	}

	schemaString, err := verbSchemaString(sch, decl.(*schema.Verb))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&pbconsole.Verb{
		Verb:              v,
		Schema:            schemaString,
		JsonRequestSchema: jsonRequestSchema,
	}), nil
}

func (c *ConsoleService) findDecl(ctx context.Context, req *pbconsole.GetDeclRequest) (schema.Decl, error) {
	deployments, err := c.dal.GetDeploymentsWithMinReplicas(ctx)
	if err != nil {
		return nil, err
	}

	module, ok := slices.Find(deployments, func(d dalmodel.Deployment) bool {
		return d.Module == req.ModuleName
	})
	if !ok {
		return nil, fmt.Errorf("could not find module with name %q", req.ModuleName)
	}
	decl, ok := slices.Find(module.Schema.Decls, func(d schema.Decl) bool {
		return d.GetName() == req.DeclName
	})
	if !ok {
		return nil, fmt.Errorf("could not find decl %q in module %q", req.DeclName, req.ModuleName)
	}
	return decl, nil
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

	results, err := c.timeline.QueryTimeline(ctx, limitPlusOne, query...)
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
			newQuery = append(newQuery, timeline.FilterTimeRange(thisRequestTime, lastEventTime))
		}

		events, err := c.timeline.QueryTimeline(ctx, int(req.Msg.Query.Limit), newQuery...)
		if err != nil {
			return err
		}

		if len(events) > 0 {
			err = stream.Send(&pbconsole.StreamEventsResponse{
				Events: slices.Map(events, eventDALToProto),
			})
			if err != nil {
				return err
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

func eventsQueryProtoToDAL(pb *pbconsole.EventsQuery) ([]timeline.TimelineFilter, error) {
	var query []timeline.TimelineFilter

	if pb.Order == pbconsole.EventsQuery_DESC {
		query = append(query, timeline.FilterDescending())
	}

	for _, filter := range pb.Filters {
		switch filter := filter.Filter.(type) {
		case *pbconsole.EventsQuery_Filter_Deployments:
			deploymentKeys := make([]model.DeploymentKey, 0, len(filter.Deployments.Deployments))
			for _, deployment := range filter.Deployments.Deployments {
				deploymentKey, err := model.ParseDeploymentKey(deployment)
				if err != nil {
					return nil, connect.NewError(connect.CodeInvalidArgument, err)
				}
				deploymentKeys = append(deploymentKeys, deploymentKey)
			}
			query = append(query, timeline.FilterDeployments(deploymentKeys...))

		case *pbconsole.EventsQuery_Filter_Requests:
			requestKeys := make([]model.RequestKey, 0, len(filter.Requests.Requests))
			for _, request := range filter.Requests.Requests {
				requestKey, err := model.ParseRequestKey(request)
				if err != nil {
					return nil, connect.NewError(connect.CodeInvalidArgument, err)
				}
				requestKeys = append(requestKeys, requestKey)
			}
			query = append(query, timeline.FilterRequests(requestKeys...))

		case *pbconsole.EventsQuery_Filter_EventTypes:
			eventTypes := make([]timeline.EventType, 0, len(filter.EventTypes.EventTypes))
			for _, eventType := range filter.EventTypes.EventTypes {
				switch eventType {
				case pbconsole.EventType_EVENT_TYPE_CALL:
					eventTypes = append(eventTypes, timeline.EventTypeCall)
				case pbconsole.EventType_EVENT_TYPE_LOG:
					eventTypes = append(eventTypes, timeline.EventTypeLog)
				case pbconsole.EventType_EVENT_TYPE_DEPLOYMENT_CREATED:
					eventTypes = append(eventTypes, timeline.EventTypeDeploymentCreated)
				case pbconsole.EventType_EVENT_TYPE_DEPLOYMENT_UPDATED:
					eventTypes = append(eventTypes, timeline.EventTypeDeploymentUpdated)
				case pbconsole.EventType_EVENT_TYPE_INGRESS:
					eventTypes = append(eventTypes, timeline.EventTypeIngress)
				default:
					return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown event type %v", eventType))
				}
			}
			query = append(query, timeline.FilterTypes(eventTypes...))

		case *pbconsole.EventsQuery_Filter_LogLevel:
			level := log.Level(filter.LogLevel.LogLevel)
			if level < log.Trace || level > log.Error {
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown log level %v", filter.LogLevel.LogLevel))
			}
			query = append(query, timeline.FilterLogLevel(level))

		case *pbconsole.EventsQuery_Filter_Time:
			var newerThan, olderThan time.Time
			if filter.Time.NewerThan != nil {
				newerThan = filter.Time.NewerThan.AsTime()
			}
			if filter.Time.OlderThan != nil {
				olderThan = filter.Time.OlderThan.AsTime()
			}
			query = append(query, timeline.FilterTimeRange(olderThan, newerThan))

		case *pbconsole.EventsQuery_Filter_Id:
			var lowerThan, higherThan int64
			if filter.Id.LowerThan != nil {
				lowerThan = *filter.Id.LowerThan
			}
			if filter.Id.HigherThan != nil {
				higherThan = *filter.Id.HigherThan
			}
			query = append(query, timeline.FilterIDRange(lowerThan, higherThan))
		case *pbconsole.EventsQuery_Filter_Call:
			var sourceModule optional.Option[string]
			if filter.Call.SourceModule != nil {
				sourceModule = optional.Some(*filter.Call.SourceModule)
			}
			var destVerb optional.Option[string]
			if filter.Call.DestVerb != nil {
				destVerb = optional.Some(*filter.Call.DestVerb)
			}
			query = append(query, timeline.FilterCall(sourceModule, filter.Call.DestModule, destVerb))

		default:
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown filter %T", filter))
		}
	}
	return query, nil
}

func eventDALToProto(event timeline.Event) *pbconsole.Event {
	switch event := event.(type) {
	case *timeline.CallEvent:
		var requestKey *string
		if r, ok := event.RequestKey.Get(); ok {
			rstr := r.String()
			requestKey = &rstr
		}
		var sourceVerbRef *schemapb.Ref
		if sourceVerb, ok := event.SourceVerb.Get(); ok {
			sourceVerbRef = sourceVerb.ToProto().(*schemapb.Ref) //nolint:forcetypeassert
		}
		return &pbconsole.Event{
			TimeStamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbconsole.Event_Call{
				Call: &pbconsole.CallEvent{
					RequestKey:    requestKey,
					DeploymentKey: event.DeploymentKey.String(),
					TimeStamp:     timestamppb.New(event.Time),
					SourceVerbRef: sourceVerbRef,
					DestinationVerbRef: &schemapb.Ref{
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

	case *timeline.LogEvent:
		var requestKey *string
		if r, ok := event.RequestKey.Get(); ok {
			rstr := r.String()
			requestKey = &rstr
		}
		return &pbconsole.Event{
			TimeStamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbconsole.Event_Log{
				Log: &pbconsole.LogEvent{
					DeploymentKey: event.DeploymentKey.String(),
					RequestKey:    requestKey,
					TimeStamp:     timestamppb.New(event.Time),
					LogLevel:      event.Level,
					Attributes:    event.Attributes,
					Message:       event.Message,
					Error:         event.Error.Ptr(),
				},
			},
		}

	case *timeline.DeploymentCreatedEvent:
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
					Key:         event.DeploymentKey.String(),
					Language:    event.Language,
					ModuleName:  event.ModuleName,
					MinReplicas: int32(event.MinReplicas),
					Replaced:    replaced,
				},
			},
		}
	case *timeline.DeploymentUpdatedEvent:
		return &pbconsole.Event{
			TimeStamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbconsole.Event_DeploymentUpdated{
				DeploymentUpdated: &pbconsole.DeploymentUpdatedEvent{
					Key:             event.DeploymentKey.String(),
					MinReplicas:     int32(event.MinReplicas),
					PrevMinReplicas: int32(event.PrevMinReplicas),
				},
			},
		}

	case *timeline.IngressEvent:
		var requestKey *string
		if r, ok := event.RequestKey.Get(); ok {
			rstr := r.String()
			requestKey = &rstr
		}

		return &pbconsole.Event{
			TimeStamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbconsole.Event_Ingress{
				Ingress: &pbconsole.IngressEvent{
					DeploymentKey: event.DeploymentKey.String(),
					RequestKey:    requestKey,
					VerbRef: &schemapb.Ref{
						Module: event.Verb.Module,
						Name:   event.Verb.Name,
					},
					Method:         event.Method,
					Path:           event.Path,
					StatusCode:     int32(event.StatusCode),
					TimeStamp:      timestamppb.New(event.Time),
					Duration:       durationpb.New(event.Duration),
					Request:        string(event.Request),
					RequestHeader:  string(event.RequestHeader),
					Response:       string(event.Response),
					ResponseHeader: string(event.ResponseHeader),
					Error:          event.Error.Ptr(),
				},
			},
		}

	default:
		panic(fmt.Errorf("unknown event type %T", event))
	}
}

func graph(sch *schema.Schema) map[string][]string {
	out := make(map[string][]string)
	for _, module := range sch.Modules {
		buildGraph(sch, module, out)
	}
	return out
}

// buildGraph recursively builds the dependency graph
func buildGraph(sch *schema.Schema, module *schema.Module, out map[string][]string) {
	out[module.Name] = module.Imports()
	for _, dep := range module.Imports() {
		var depModule *schema.Module
		for _, m := range sch.Modules {
			if m.String() == dep {
				depModule = m
				break
			}
		}
		if depModule != nil {
			buildGraph(sch, module, out)
		}
	}
}
