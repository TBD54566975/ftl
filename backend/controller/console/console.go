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
	"github.com/TBD54566975/ftl/internal/buildengine"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
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
				verb, err := verbFromDecl(decl, sch)
				if err != nil {
					return nil, err
				}
				verbs = append(verbs, verb)

			case *schema.Data:
				data = append(data, dataFromDecl(decl))

			case *schema.Secret:
				secrets = append(secrets, secretFromDecl(decl))

			case *schema.Config:
				configs = append(configs, configFromDecl(decl))

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

func moduleFromDeployment(deployment dalmodel.Deployment, sch *schema.Schema) (*pbconsole.Module, error) {
	module, err := moduleFromDecls(deployment.Schema.Decls, sch)
	if err != nil {
		return nil, err
	}

	module.Name = deployment.Module
	module.DeploymentKey = deployment.Key.String()
	module.Language = deployment.Language
	module.Schema = deployment.Schema.String()

	return module, nil
}

func moduleFromDecls(decls []schema.Decl, sch *schema.Schema) (*pbconsole.Module, error) {
	var configs []*pbconsole.Config
	var data []*pbconsole.Data
	var databases []*pbconsole.Database
	var enums []*pbconsole.Enum
	var fsms []*pbconsole.FSM
	var topics []*pbconsole.Topic
	var typealiases []*pbconsole.TypeAlias
	var secrets []*pbconsole.Secret
	var subscriptions []*pbconsole.Subscription
	var verbs []*pbconsole.Verb

	for _, decl := range decls {
		switch decl := decl.(type) {
		case *schema.Config:
			configs = append(configs, configFromDecl(decl))

		case *schema.Data:
			data = append(data, dataFromDecl(decl))

		case *schema.Database:
			databases = append(databases, databaseFromDecl(decl))

		case *schema.Enum:
			enums = append(enums, enumFromDecl(decl))

		case *schema.FSM:
			fsms = append(fsms, fsmFromDecl(decl))

		case *schema.Topic:
			topics = append(topics, topicFromDecl(decl))

		case *schema.Secret:
			secrets = append(secrets, secretFromDecl(decl))

		case *schema.Subscription:
			subscriptions = append(subscriptions, subscriptionFromDecl(decl))

		case *schema.TypeAlias:
			typealiases = append(typealiases, typealiasFromDecl(decl))

		case *schema.Verb:
			verb, err := verbFromDecl(decl, sch)
			if err != nil {
				return nil, err
			}
			verbs = append(verbs, verb)
		}
	}

	return &pbconsole.Module{
		Configs:       configs,
		Data:          data,
		Databases:     databases,
		Enums:         enums,
		Fsms:          fsms,
		Topics:        topics,
		Typealiases:   typealiases,
		Secrets:       secrets,
		Subscriptions: subscriptions,
		Verbs:         verbs,
	}, nil
}

func configFromDecl(decl *schema.Config) *pbconsole.Config {
	//nolint:forcetypeassert
	config := decl.ToProto().(*schemapb.Config)
	return &pbconsole.Config{
		Config: config,
	}
}

func dataFromDecl(decl *schema.Data) *pbconsole.Data {
	//nolint:forcetypeassert
	d := decl.ToProto().(*schemapb.Data)
	return &pbconsole.Data{
		Data:   d,
		Schema: schema.DataFromProto(d).String(),
	}
}

func databaseFromDecl(decl *schema.Database) *pbconsole.Database {
	//nolint:forcetypeassert
	d := decl.ToProto().(*schemapb.Database)
	return &pbconsole.Database{
		Database: d,
	}
}

func enumFromDecl(decl *schema.Enum) *pbconsole.Enum {
	//nolint:forcetypeassert
	e := decl.ToProto().(*schemapb.Enum)
	return &pbconsole.Enum{
		Enum: e,
	}
}

func fsmFromDecl(decl *schema.FSM) *pbconsole.FSM {
	//nolint:forcetypeassert
	f := decl.ToProto().(*schemapb.FSM)
	return &pbconsole.FSM{
		Fsm: f,
	}
}

func topicFromDecl(decl *schema.Topic) *pbconsole.Topic {
	//nolint:forcetypeassert
	t := decl.ToProto().(*schemapb.Topic)
	return &pbconsole.Topic{
		Topic: t,
	}
}

func typealiasFromDecl(decl *schema.TypeAlias) *pbconsole.TypeAlias {
	//nolint:forcetypeassert
	t := decl.ToProto().(*schemapb.TypeAlias)
	return &pbconsole.TypeAlias{
		Typealias: t,
	}
}

func secretFromDecl(decl *schema.Secret) *pbconsole.Secret {
	//nolint:forcetypeassert
	s := decl.ToProto().(*schemapb.Secret)
	return &pbconsole.Secret{
		Secret: s,
	}
}

func subscriptionFromDecl(decl *schema.Subscription) *pbconsole.Subscription {
	//nolint:forcetypeassert
	s := decl.ToProto().(*schemapb.Subscription)
	return &pbconsole.Subscription{
		Subscription: s,
	}
}

func verbFromDecl(decl *schema.Verb, sch *schema.Schema) (*pbconsole.Verb, error) {
	//nolint:forcetypeassert
	v := decl.ToProto().(*schemapb.Verb)
	verbSchema := schema.VerbFromProto(v)
	var jsonRequestSchema string
	if verbSchema.Request != nil {
		if requestData, ok := verbSchema.Request.(*schema.Ref); ok {
			jsonSchema, err := schema.RequestResponseToJSONSchema(sch, *requestData)
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve JSON schema: %w", err)
			}
			jsonData, err := json.MarshalIndent(jsonSchema, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to indent JSON schema: %w", err)
			}
			jsonRequestSchema = string(jsonData)
		}
	}

	schemaString, err := verbSchemaString(sch, decl)
	if err != nil {
		return nil, err
	}
	return &pbconsole.Verb{
		Verb:              v,
		Schema:            schemaString,
		JsonRequestSchema: jsonRequestSchema,
	}, nil
}

func (c *ConsoleService) StreamModules(ctx context.Context, req *connect.Request[pbconsole.StreamModulesRequest], stream *connect.ServerStream[pbconsole.StreamModulesResponse]) error {
	deployments, err := c.dal.GetDeploymentsWithMinReplicas(ctx)
	if err != nil {
		return fmt.Errorf("failed to get deployments: %w", err)
	}
	sch := &schema.Schema{
		Modules: slices.Map(deployments, func(d dalmodel.Deployment) *schema.Module {
			return d.Schema
		}),
	}
	builtin := schema.Builtins()
	sch.Modules = append(sch.Modules, builtin)

	var modules []*pbconsole.Module
	for _, deployment := range deployments {
		module, err := moduleFromDeployment(deployment, sch)
		if err != nil {
			return err
		}
		modules = append(modules, module)
	}

	builtinModule, err := moduleFromDecls(builtin.Decls, sch)
	if err != nil {
		return err
	}
	builtinModule.Name = builtin.Name
	builtinModule.Language = "go"
	builtinModule.Schema = builtin.String()
	modules = append(modules, builtinModule)

	err = stream.Send(&pbconsole.StreamModulesResponse{
		Modules: modules,
	})
	if err != nil {
		return fmt.Errorf("failed to send StreamModulesResponse to stream: %w", err)
	}

	// TODO: handle deployment updates
	return nil
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
				case pbconsole.EventType_EVENT_TYPE_CRON_SCHEDULED:
					eventTypes = append(eventTypes, timeline.EventTypeCronScheduled)
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
		case *pbconsole.EventsQuery_Filter_Module:
			var verb optional.Option[string]
			if filter.Module.Verb != nil {
				verb = optional.Some(*filter.Module.Verb)
			}
			query = append(query, timeline.FilterModule(filter.Module.Module, verb))

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

	case *timeline.CronScheduledEvent:
		return &pbconsole.Event{
			TimeStamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbconsole.Event_CronScheduled{
				CronScheduled: &pbconsole.CronScheduledEvent{
					DeploymentKey: event.DeploymentKey.String(),
					VerbRef: &schemapb.Ref{
						Module: event.Verb.Module,
						Name:   event.Verb.Name,
					},
					TimeStamp:   timestamppb.New(event.Time),
					Duration:    durationpb.New(event.Duration),
					ScheduledAt: timestamppb.New(event.ScheduledAt),
					Schedule:    event.Schedule,
					Error:       event.Error.Ptr(),
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
