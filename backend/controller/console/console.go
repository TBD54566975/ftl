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

	"github.com/TBD54566975/ftl/backend/controller/admin"
	"github.com/TBD54566975/ftl/backend/controller/dal"
	dalmodel "github.com/TBD54566975/ftl/backend/controller/dal/model"
	"github.com/TBD54566975/ftl/backend/controller/timeline"
	pbconsole "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/console/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/console/v1/pbconsoleconnect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	pbtimeline "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/buildengine"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/slices"
)

type ConsoleService struct {
	dal      *dal.DAL
	timeline *timeline.Service
	admin    *admin.AdminService
}

var _ pbconsoleconnect.ConsoleServiceHandler = (*ConsoleService)(nil)

func NewService(dal *dal.DAL, timeline *timeline.Service, admin *admin.AdminService) *ConsoleService {
	return &ConsoleService{
		dal:      dal,
		timeline: timeline,
		admin:    admin,
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

	nilMap := map[schema.RefKey]map[schema.RefKey]bool{}
	var modules []*pbconsole.Module
	for _, deployment := range deployments {
		var verbs []*pbconsole.Verb
		var data []*pbconsole.Data
		var secrets []*pbconsole.Secret
		var configs []*pbconsole.Config

		for _, decl := range deployment.Schema.Decls {
			switch decl := decl.(type) {
			case *schema.Verb:
				verb, err := verbFromDecl(decl, sch, deployment.Module, nilMap)
				if err != nil {
					return nil, err
				}
				verbs = append(verbs, verb)

			case *schema.Data:
				data = append(data, dataFromDecl(decl, deployment.Module, nilMap))

			case *schema.Secret:
				secrets = append(secrets, secretFromDecl(decl, deployment.Module, nilMap))

			case *schema.Config:
				configs = append(configs, configFromDecl(decl, deployment.Module, nilMap))

			case *schema.Database, *schema.Enum, *schema.TypeAlias, *schema.Topic:
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

func moduleFromDeployment(deployment dalmodel.Deployment, sch *schema.Schema, refMap map[schema.RefKey]map[schema.RefKey]bool) (*pbconsole.Module, error) {
	module, err := moduleFromDecls(deployment.Schema.Decls, sch, deployment.Module, refMap)
	if err != nil {
		return nil, err
	}

	module.Name = deployment.Module
	module.DeploymentKey = deployment.Key.String()
	module.Language = deployment.Language
	module.Schema = deployment.Schema.String()

	return module, nil
}

func moduleFromDecls(decls []schema.Decl, sch *schema.Schema, module string, refMap map[schema.RefKey]map[schema.RefKey]bool) (*pbconsole.Module, error) {
	var configs []*pbconsole.Config
	var data []*pbconsole.Data
	var databases []*pbconsole.Database
	var enums []*pbconsole.Enum
	var topics []*pbconsole.Topic
	var typealiases []*pbconsole.TypeAlias
	var secrets []*pbconsole.Secret
	var verbs []*pbconsole.Verb

	for _, d := range decls {
		switch decl := d.(type) {
		case *schema.Config:
			configs = append(configs, configFromDecl(decl, module, refMap))

		case *schema.Data:
			data = append(data, dataFromDecl(decl, module, refMap))

		case *schema.Database:
			databases = append(databases, databaseFromDecl(decl, module, refMap))

		case *schema.Enum:
			enums = append(enums, enumFromDecl(decl, module, refMap))

		case *schema.Topic:
			topics = append(topics, topicFromDecl(decl, module, refMap))

		case *schema.Secret:
			secrets = append(secrets, secretFromDecl(decl, module, refMap))

		case *schema.TypeAlias:
			typealiases = append(typealiases, typealiasFromDecl(decl, module, refMap))

		case *schema.Verb:
			verb, err := verbFromDecl(decl, sch, module, refMap)
			if err != nil {
				return nil, err
			}
			verbs = append(verbs, verb)
		}
	}

	return &pbconsole.Module{
		Configs:     configs,
		Data:        data,
		Databases:   databases,
		Enums:       enums,
		Topics:      topics,
		Typealiases: typealiases,
		Secrets:     secrets,
		Verbs:       verbs,
	}, nil
}

func configFromDecl(decl *schema.Config, module string, refMap map[schema.RefKey]map[schema.RefKey]bool) *pbconsole.Config {
	return &pbconsole.Config{
		//nolint:forcetypeassert
		Config:     decl.ToProto().(*schemapb.Config),
		References: getReferencesFromMap(refMap, module, decl.Name),
	}
}

func dataFromDecl(decl *schema.Data, module string, refMap map[schema.RefKey]map[schema.RefKey]bool) *pbconsole.Data {
	//nolint:forcetypeassert
	d := decl.ToProto().(*schemapb.Data)
	return &pbconsole.Data{
		Data:       d,
		Schema:     schema.DataFromProto(d).String(),
		References: getReferencesFromMap(refMap, module, decl.Name),
	}
}

func databaseFromDecl(decl *schema.Database, module string, refMap map[schema.RefKey]map[schema.RefKey]bool) *pbconsole.Database {
	return &pbconsole.Database{
		//nolint:forcetypeassert
		Database:   decl.ToProto().(*schemapb.Database),
		References: getReferencesFromMap(refMap, module, decl.Name),
	}
}

func enumFromDecl(decl *schema.Enum, module string, refMap map[schema.RefKey]map[schema.RefKey]bool) *pbconsole.Enum {
	return &pbconsole.Enum{
		//nolint:forcetypeassert
		Enum:       decl.ToProto().(*schemapb.Enum),
		References: getReferencesFromMap(refMap, module, decl.Name),
	}
}

func topicFromDecl(decl *schema.Topic, module string, refMap map[schema.RefKey]map[schema.RefKey]bool) *pbconsole.Topic {
	return &pbconsole.Topic{
		//nolint:forcetypeassert
		Topic:      decl.ToProto().(*schemapb.Topic),
		References: getReferencesFromMap(refMap, module, decl.Name),
	}
}

func typealiasFromDecl(decl *schema.TypeAlias, module string, refMap map[schema.RefKey]map[schema.RefKey]bool) *pbconsole.TypeAlias {
	return &pbconsole.TypeAlias{
		//nolint:forcetypeassert
		Typealias:  decl.ToProto().(*schemapb.TypeAlias),
		References: getReferencesFromMap(refMap, module, decl.Name),
	}
}

func secretFromDecl(decl *schema.Secret, module string, refMap map[schema.RefKey]map[schema.RefKey]bool) *pbconsole.Secret {
	return &pbconsole.Secret{
		//nolint:forcetypeassert
		Secret:     decl.ToProto().(*schemapb.Secret),
		References: getReferencesFromMap(refMap, module, decl.Name),
	}
}

func verbFromDecl(decl *schema.Verb, sch *schema.Schema, module string, refMap map[schema.RefKey]map[schema.RefKey]bool) (*pbconsole.Verb, error) {
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
		References:        getReferencesFromMap(refMap, module, decl.Name),
	}, nil
}

func getReferencesFromMap(refMap map[schema.RefKey]map[schema.RefKey]bool, module string, name string) []*schemapb.Ref {
	key := schema.RefKey{
		Module: module,
		Name:   name,
	}
	out := []*schemapb.Ref{}
	if refs, ok := refMap[key]; ok {
		for refKey, ok := range refs {
			if ok {
				out = append(out, refKey.ToProto())
			}
		}
	}
	return out
}

func (c *ConsoleService) StreamModules(ctx context.Context, req *connect.Request[pbconsole.StreamModulesRequest], stream *connect.ServerStream[pbconsole.StreamModulesResponse]) error {
	deploymentChanges := make(chan dal.DeploymentNotification, 32)

	// Subscribe to deployment changes.
	c.dal.DeploymentChanges.Subscribe(deploymentChanges)
	defer c.dal.DeploymentChanges.Unsubscribe(deploymentChanges)

	err := c.sendStreamModulesResp(ctx, stream)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-deploymentChanges:
			err = c.sendStreamModulesResp(ctx, stream)
			if err != nil {
				return err
			}
		}
	}
}

// filterDeployments removes any duplicate modules by selecting the deployment with the
// latest CreatedAt.
func (c *ConsoleService) filterDeployments(unfilteredDeployments []dalmodel.Deployment) []dalmodel.Deployment {
	latest := make(map[string]dalmodel.Deployment)

	for _, deployment := range unfilteredDeployments {
		if existing, found := latest[deployment.Module]; !found || deployment.CreatedAt.After(existing.CreatedAt) {
			latest[deployment.Module] = deployment

		}
	}

	var result []dalmodel.Deployment
	for _, value := range latest {
		result = append(result, value)
	}

	return result
}

func (c *ConsoleService) sendStreamModulesResp(ctx context.Context, stream *connect.ServerStream[pbconsole.StreamModulesResponse]) error {
	unfilteredDeployments, err := c.dal.GetDeploymentsWithMinReplicas(ctx)
	if err != nil {
		return fmt.Errorf("failed to get deployments: %w", err)
	}
	deployments := c.filterDeployments(unfilteredDeployments)
	sch := &schema.Schema{
		Modules: slices.Map(deployments, func(d dalmodel.Deployment) *schema.Module {
			return d.Schema
		}),
	}
	builtin := schema.Builtins()
	sch.Modules = append(sch.Modules, builtin)

	// Get topology
	sorted, err := buildengine.TopologicalSort(graph(sch))
	if err != nil {
		return fmt.Errorf("failed to sort modules: %w", err)
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

	refMap, err := getSchemaRefs(sch)
	if err != nil {
		return fmt.Errorf("failed to find references: %w", err)
	}

	var modules []*pbconsole.Module
	for _, deployment := range deployments {
		module, err := moduleFromDeployment(deployment, sch, refMap)
		if err != nil {
			return err
		}
		modules = append(modules, module)
	}

	builtinModule, err := moduleFromDecls(builtin.Decls, sch, builtin.Name, refMap)
	if err != nil {
		return err
	}
	builtinModule.Name = builtin.Name
	builtinModule.Language = "go"
	builtinModule.Schema = builtin.String()
	modules = append(modules, builtinModule)

	err = stream.Send(&pbconsole.StreamModulesResponse{
		Modules:  modules,
		Topology: topology,
	})
	if err != nil {
		return fmt.Errorf("failed to send StreamModulesResponse to stream: %w", err)
	}

	return nil
}

func getSchemaRefs(sch *schema.Schema) (map[schema.RefKey]map[schema.RefKey]bool, error) {
	refsToReferers := map[schema.RefKey]map[schema.RefKey]bool{}
	for _, module := range sch.Modules {
		for _, parentDecl := range module.Decls {
			parentDeclRef := schema.Ref{
				Module: module.Name,
				Name:   parentDecl.GetName(),
			}
			err := schema.Visit(parentDecl, func(n schema.Node, next func() error) error {
				if ref, ok := n.(*schema.Ref); ok {
					addRefToSetMap(refsToReferers, ref.ToRefKey(), parentDeclRef)
				}
				return next()
			})
			if err != nil {
				return nil, fmt.Errorf("visit failed: %w", err)
			}
		}
	}
	return refsToReferers, nil
}

// addRefToSetMap approximates adding to a map[ref]->set[ref], where the "set" is implemented
// as a map to bools. A value is in the set if its value is `true`.
func addRefToSetMap(m map[schema.RefKey]map[schema.RefKey]bool, key schema.RefKey, value schema.Ref) {
	_, ok := m[key]
	if !ok {
		m[key] = map[schema.RefKey]bool{}
	}
	m[key][value.ToRefKey()] = true
}

func (c *ConsoleService) GetEvents(ctx context.Context, req *connect.Request[pbconsole.GetEventsRequest]) (*connect.Response[pbconsole.GetEventsResponse], error) {
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

//nolint:maintidx
func eventsQueryProtoToDAL(query *pbconsole.GetEventsRequest) ([]timeline.TimelineFilter, error) {
	var result []timeline.TimelineFilter

	if query.Order == pbconsole.GetEventsRequest_ORDER_DESC {
		result = append(result, timeline.FilterDescending())
	}

	for _, filter := range query.Filters {
		switch f := filter.Filter.(type) {
		case *pbconsole.GetEventsRequest_Filter_Deployments:
			deploymentKeys := make([]model.DeploymentKey, 0, len(f.Deployments.Deployments))
			for _, deployment := range f.Deployments.Deployments {
				deploymentKey, err := model.ParseDeploymentKey(deployment)
				if err != nil {
					return nil, connect.NewError(connect.CodeInvalidArgument, err)
				}
				deploymentKeys = append(deploymentKeys, deploymentKey)
			}
			result = append(result, timeline.FilterDeployments(deploymentKeys...))

		case *pbconsole.GetEventsRequest_Filter_Requests:
			requestKeys := make([]model.RequestKey, 0, len(f.Requests.Requests))
			for _, request := range f.Requests.Requests {
				requestKey, err := model.ParseRequestKey(request)
				if err != nil {
					return nil, connect.NewError(connect.CodeInvalidArgument, err)
				}
				requestKeys = append(requestKeys, requestKey)
			}
			result = append(result, timeline.FilterRequests(requestKeys...))

		case *pbconsole.GetEventsRequest_Filter_EventTypes:
			var types []timeline.EventType
			for _, t := range f.EventTypes.EventTypes {
				types = append(types, timeline.EventType(t))
			}
			result = append(result, timeline.FilterTypes(types...))

		case *pbconsole.GetEventsRequest_Filter_LogLevel:
			level := log.Level(f.LogLevel.LogLevel)
			if level < log.Trace || level > log.Error {
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown log level %v", f.LogLevel.LogLevel))
			}
			result = append(result, timeline.FilterLogLevel(level))

		case *pbconsole.GetEventsRequest_Filter_Time:
			newerThan := f.Time.GetNewerThan().AsTime()
			olderThan := f.Time.GetOlderThan().AsTime()
			result = append(result, timeline.FilterTimeRange(olderThan, newerThan))

		case *pbconsole.GetEventsRequest_Filter_Id:
			lowerThan := f.Id.GetLowerThan()
			higherThan := f.Id.GetHigherThan()
			result = append(result, timeline.FilterIDRange(lowerThan, higherThan))

		case *pbconsole.GetEventsRequest_Filter_Call:
			sourceModule := optional.Zero(f.Call.GetSourceModule())
			destVerb := optional.Zero(f.Call.GetDestVerb())
			result = append(result, timeline.FilterCall(sourceModule, f.Call.DestModule, destVerb))

		default:
			return nil, fmt.Errorf("unknown filter type %T", f)
		}
	}

	return result, nil
}

//nolint:maintidx
func eventDALToProto(event timeline.Event) *pbtimeline.Event {
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
		return &pbtimeline.Event{
			Timestamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbtimeline.Event_Call{
				Call: &pbtimeline.CallEvent{
					RequestKey:    requestKey,
					DeploymentKey: event.DeploymentKey.String(),
					Timestamp:     timestamppb.New(event.Time),
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
		return &pbtimeline.Event{
			Timestamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbtimeline.Event_Log{
				Log: &pbtimeline.LogEvent{
					DeploymentKey: event.DeploymentKey.String(),
					RequestKey:    requestKey,
					Timestamp:     timestamppb.New(event.Time),
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
		return &pbtimeline.Event{
			Timestamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbtimeline.Event_DeploymentCreated{
				DeploymentCreated: &pbtimeline.DeploymentCreatedEvent{
					Key:         event.DeploymentKey.String(),
					Language:    event.Language,
					ModuleName:  event.ModuleName,
					MinReplicas: int32(event.MinReplicas),
					Replaced:    replaced,
				},
			},
		}
	case *timeline.DeploymentUpdatedEvent:
		return &pbtimeline.Event{
			Timestamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbtimeline.Event_DeploymentUpdated{
				DeploymentUpdated: &pbtimeline.DeploymentUpdatedEvent{
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

		return &pbtimeline.Event{
			Timestamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbtimeline.Event_Ingress{
				Ingress: &pbtimeline.IngressEvent{
					DeploymentKey: event.DeploymentKey.String(),
					RequestKey:    requestKey,
					VerbRef: &schemapb.Ref{
						Module: event.Verb.Module,
						Name:   event.Verb.Name,
					},
					Method:         event.Method,
					Path:           event.Path,
					StatusCode:     int32(event.StatusCode),
					Timestamp:      timestamppb.New(event.Time),
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
		return &pbtimeline.Event{
			Timestamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbtimeline.Event_CronScheduled{
				CronScheduled: &pbtimeline.CronScheduledEvent{
					DeploymentKey: event.DeploymentKey.String(),
					VerbRef: &schemapb.Ref{
						Module: event.Verb.Module,
						Name:   event.Verb.Name,
					},
					Timestamp:   timestamppb.New(event.Time),
					Duration:    durationpb.New(event.Duration),
					ScheduledAt: timestamppb.New(event.ScheduledAt),
					Schedule:    event.Schedule,
					Error:       event.Error.Ptr(),
				},
			},
		}

	case *timeline.AsyncExecuteEvent:
		var requestKey *string
		if rstr, ok := event.RequestKey.Get(); ok {
			requestKey = &rstr
		}

		var asyncEventType pbtimeline.AsyncExecuteEventType
		switch event.EventType {
		case timeline.AsyncExecuteEventTypeUnkown:
			asyncEventType = pbtimeline.AsyncExecuteEventType_ASYNC_EXECUTE_EVENT_TYPE_UNSPECIFIED
		case timeline.AsyncExecuteEventTypeCron:
			asyncEventType = pbtimeline.AsyncExecuteEventType_ASYNC_EXECUTE_EVENT_TYPE_CRON
		case timeline.AsyncExecuteEventTypePubSub:
			asyncEventType = pbtimeline.AsyncExecuteEventType_ASYNC_EXECUTE_EVENT_TYPE_PUBSUB
		}

		return &pbtimeline.Event{
			Timestamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbtimeline.Event_AsyncExecute{
				AsyncExecute: &pbtimeline.AsyncExecuteEvent{
					DeploymentKey:  event.DeploymentKey.String(),
					RequestKey:     requestKey,
					Timestamp:      timestamppb.New(event.Time),
					AsyncEventType: asyncEventType,
					VerbRef: &schemapb.Ref{
						Module: event.Verb.Module,
						Name:   event.Verb.Name,
					},
					Duration: durationpb.New(event.Duration),
					Error:    event.Error.Ptr(),
				},
			},
		}

	case *timeline.PubSubPublishEvent:
		var requestKey *string
		if r, ok := event.RequestKey.Get(); ok {
			requestKey = &r
		}

		return &pbtimeline.Event{
			Timestamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbtimeline.Event_PubsubPublish{
				PubsubPublish: &pbtimeline.PubSubPublishEvent{
					DeploymentKey: event.DeploymentKey.String(),
					RequestKey:    requestKey,
					VerbRef:       event.SourceVerb.ToProto().(*schemapb.Ref), //nolint:forcetypeassert
					Timestamp:     timestamppb.New(event.Time),
					Duration:      durationpb.New(event.Duration),
					Topic:         event.Topic,
					Request:       string(event.Request),
					Error:         event.Error.Ptr(),
				},
			},
		}

	case *timeline.PubSubConsumeEvent:
		var requestKey *string
		if r, ok := event.RequestKey.Get(); ok {
			requestKey = &r
		}

		var destVerbModule string
		var destVerbName string
		if destVerb, ok := event.DestVerb.Get(); ok {
			destVerbModule = destVerb.Module
			destVerbName = destVerb.Name
		}

		return &pbtimeline.Event{
			Timestamp: timestamppb.New(event.Time),
			Id:        event.ID,
			Entry: &pbtimeline.Event_PubsubConsume{
				PubsubConsume: &pbtimeline.PubSubConsumeEvent{
					DeploymentKey:  event.DeploymentKey.String(),
					RequestKey:     requestKey,
					DestVerbModule: &destVerbModule,
					DestVerbName:   &destVerbName,
					Timestamp:      timestamppb.New(event.Time),
					Duration:       durationpb.New(event.Duration),
					Topic:          event.Topic,
					Error:          event.Error.Ptr(),
				},
			},
		}

	default:
		panic(fmt.Errorf("unknown event type %T", event))
	}
}

func (c *ConsoleService) GetConfig(ctx context.Context, req *connect.Request[pbconsole.GetConfigRequest]) (*connect.Response[pbconsole.GetConfigResponse], error) {
	resp, err := c.admin.ConfigGet(ctx, connect.NewRequest(&ftlv1.ConfigGetRequest{
		Ref: &ftlv1.ConfigRef{
			Module: req.Msg.Module,
			Name:   req.Msg.Name,
		},
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	return connect.NewResponse(&pbconsole.GetConfigResponse{
		Value: resp.Msg.Value,
	}), nil
}

func (c *ConsoleService) SetConfig(ctx context.Context, req *connect.Request[pbconsole.SetConfigRequest]) (*connect.Response[pbconsole.SetConfigResponse], error) {
	_, err := c.admin.ConfigSet(ctx, connect.NewRequest(&ftlv1.ConfigSetRequest{
		Ref: &ftlv1.ConfigRef{
			Module: req.Msg.Module,
			Name:   req.Msg.Name,
		},
		Value: req.Msg.Value,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to set config: %w", err)
	}
	return connect.NewResponse(&pbconsole.SetConfigResponse{}), nil
}

func (c *ConsoleService) GetSecret(ctx context.Context, req *connect.Request[pbconsole.GetSecretRequest]) (*connect.Response[pbconsole.GetSecretResponse], error) {
	resp, err := c.admin.SecretGet(ctx, connect.NewRequest(&ftlv1.SecretGetRequest{
		Ref: &ftlv1.ConfigRef{
			Name:   req.Msg.Name,
			Module: req.Msg.Module,
		},
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}
	return connect.NewResponse(&pbconsole.GetSecretResponse{
		Value: resp.Msg.Value,
	}), nil
}

func (c *ConsoleService) SetSecret(ctx context.Context, req *connect.Request[pbconsole.SetSecretRequest]) (*connect.Response[pbconsole.SetSecretResponse], error) {
	_, err := c.admin.SecretSet(ctx, connect.NewRequest(&ftlv1.SecretSetRequest{
		Ref: &ftlv1.ConfigRef{
			Name:   req.Msg.Name,
			Module: req.Msg.Module,
		},
		Value: req.Msg.Value,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to set secret: %w", err)
	}
	return connect.NewResponse(&pbconsole.SetSecretResponse{}), nil
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
