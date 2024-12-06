package console

import (
	"context"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"

	"github.com/TBD54566975/ftl/backend/controller/admin"
	"github.com/TBD54566975/ftl/backend/controller/dal"
	dalmodel "github.com/TBD54566975/ftl/backend/controller/dal/model"
	pbconsole "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/console/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/console/v1/pbconsoleconnect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	timelinepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1/timelinev1connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/buildengine"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/slices"
)

type ConsoleService struct {
	dal   *dal.DAL
	admin *admin.AdminService
}

var _ pbconsoleconnect.ConsoleServiceHandler = (*ConsoleService)(nil)
var _ timelinev1connect.TimelineServiceHandler = (*ConsoleService)(nil)

func NewService(dal *dal.DAL, admin *admin.AdminService) *ConsoleService {
	return &ConsoleService{
		dal:   dal,
		admin: admin,
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

func (c *ConsoleService) GetTimeline(ctx context.Context, req *connect.Request[timelinepb.GetTimelineRequest]) (*connect.Response[timelinepb.GetTimelineResponse], error) {
	client := rpc.ClientFromContext[timelinev1connect.TimelineServiceClient](ctx)
	resp, err := client.GetTimeline(ctx, connect.NewRequest(req.Msg))
	if err != nil {
		return nil, fmt.Errorf("failed to get timeline from service: %w", err)
	}
	return connect.NewResponse(resp.Msg), nil
}

func (c *ConsoleService) StreamTimeline(ctx context.Context, req *connect.Request[timelinepb.StreamTimelineRequest], out *connect.ServerStream[timelinepb.StreamTimelineResponse]) error {
	client := rpc.ClientFromContext[timelinev1connect.TimelineServiceClient](ctx)
	stream, err := client.StreamTimeline(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to stream timeline from service: %w", err)
	}
	defer stream.Close()
	for stream.Receive() {
		msg := stream.Msg()
		err = out.Send(msg)
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	}
	if stream.Err() != nil {
		return fmt.Errorf("error streaming timeline from service: %w", stream.Err())
	}
	return nil
}
func (c *ConsoleService) CreateEvent(ctx context.Context, req *connect.Request[timelinepb.CreateEventRequest]) (*connect.Response[timelinepb.CreateEventResponse], error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *ConsoleService) DeleteOldEvents(ctx context.Context, req *connect.Request[timelinepb.DeleteOldEventsRequest]) (*connect.Response[timelinepb.DeleteOldEventsResponse], error) {
	return nil, fmt.Errorf("not implemented")
}
