package console

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"connectrpc.com/connect"

	"github.com/block/ftl/backend/admin"
	consolepb "github.com/block/ftl/backend/protos/xyz/block/ftl/console/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/console/v1/consolepbconnect"
	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/timeline"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	frontend "github.com/block/ftl/frontend/console"
	"github.com/block/ftl/internal/buildengine"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type Config struct {
	ConsoleURL  *url.URL  `help:"The public URL of the console (for CORS)." env:"FTL_CONTROLLER_CONSOLE_URL"`
	ContentTime time.Time `help:"Time to use for console resource timestamps." default:"${timestamp=1970-01-01T00:00:00Z}"`
	Bind        *url.URL  `help:"Socket to bind to." default:"http://127.0.0.1:8899" env:"FTL_BIND"`
}

type service struct {
	schemaEventSource schemaeventsource.EventSource
	controllerClient  ftlv1connect.ControllerServiceClient
	timelineClient    *timeline.Client
	adminClient       admin.Client
	callClient        routing.CallClient
}

var _ consolepbconnect.ConsoleServiceHandler = (*service)(nil)

func Start(ctx context.Context, config Config, eventSource schemaeventsource.EventSource, controllerClient ftlv1connect.ControllerServiceClient, timelineClient *timeline.Client, adminClient admin.Client, client routing.CallClient) error {
	logger := log.FromContext(ctx).Scope("console")
	ctx = log.ContextWithLogger(ctx, logger)

	svc := &service{
		schemaEventSource: eventSource,
		controllerClient:  controllerClient,
		timelineClient:    timelineClient,
		adminClient:       adminClient,
		callClient:        client,
	}

	consoleHandler, err := frontend.Server(ctx, config.ContentTime, config.Bind)
	if err != nil {
		return fmt.Errorf("could not start console: %w", err)
	}
	logger.Infof("Web console available at: %s", config.Bind)

	logger.Debugf("Console service listening on: %s", config.Bind)
	err = rpc.Serve(ctx, config.Bind,
		rpc.GRPC(consolepbconnect.NewConsoleServiceHandler, svc),
		rpc.HTTP("/", consoleHandler),
	)
	if err != nil {
		return fmt.Errorf("console service stopped serving: %w", err)
	}
	return nil
}

func (s *service) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
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

func (s *service) GetModules(ctx context.Context, req *connect.Request[consolepb.GetModulesRequest]) (*connect.Response[consolepb.GetModulesResponse], error) {
	sch := s.schemaEventSource.View()

	nilMap := map[schema.RefKey]map[schema.RefKey]bool{}
	var modules []*consolepb.Module
	for _, mod := range sch.Modules {
		if mod.Runtime == nil || mod.Runtime.Deployment == nil {
			continue
		}
		var verbs []*consolepb.Verb
		var data []*consolepb.Data
		var secrets []*consolepb.Secret
		var configs []*consolepb.Config

		for _, decl := range mod.Decls {
			switch decl := decl.(type) {
			case *schema.Verb:
				verb, err := verbFromDecl(decl, sch, mod.Name, nilMap)
				if err != nil {
					return nil, err
				}
				verbs = append(verbs, verb)

			case *schema.Data:
				data = append(data, dataFromDecl(decl, mod.Name, nilMap))

			case *schema.Secret:
				secrets = append(secrets, secretFromDecl(decl, mod.Name, nilMap))

			case *schema.Config:
				configs = append(configs, configFromDecl(decl, mod.Name, nilMap))

			case *schema.Database, *schema.Enum, *schema.TypeAlias, *schema.Topic:
			}
		}

		modules = append(modules, &consolepb.Module{
			Name:          mod.Name,
			DeploymentKey: mod.Runtime.Deployment.DeploymentKey,
			Language:      mod.Runtime.Base.Language,
			Verbs:         verbs,
			Data:          data,
			Secrets:       secrets,
			Configs:       configs,
			Schema:        mod.String(),
		})
	}

	sorted, err := buildengine.TopologicalSort(graph(sch))
	if err != nil {
		return nil, fmt.Errorf("failed to sort modules: %w", err)
	}
	topology := &consolepb.Topology{
		Levels: make([]*consolepb.TopologyGroup, len(sorted)),
	}
	for i, level := range sorted {
		group := &consolepb.TopologyGroup{
			Modules: level,
		}
		topology.Levels[i] = group
	}

	return connect.NewResponse(&consolepb.GetModulesResponse{
		Modules:  modules,
		Topology: topology,
	}), nil
}

func moduleFromDeployment(deployment *schema.Module, sch *schema.Schema, refMap map[schema.RefKey]map[schema.RefKey]bool) (*consolepb.Module, error) {
	module, err := moduleFromDecls(deployment.Decls, sch, deployment.Name, refMap)
	if err != nil {
		return nil, err
	}

	module.Name = deployment.Name
	module.DeploymentKey = deployment.Runtime.Deployment.DeploymentKey
	module.Language = deployment.Runtime.Base.Language
	module.Schema = deployment.String()

	return module, nil
}

func moduleFromDecls(decls []schema.Decl, sch *schema.Schema, module string, refMap map[schema.RefKey]map[schema.RefKey]bool) (*consolepb.Module, error) {
	var configs []*consolepb.Config
	var data []*consolepb.Data
	var databases []*consolepb.Database
	var enums []*consolepb.Enum
	var topics []*consolepb.Topic
	var typealiases []*consolepb.TypeAlias
	var secrets []*consolepb.Secret
	var verbs []*consolepb.Verb

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

	return &consolepb.Module{
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

func configFromDecl(decl *schema.Config, module string, refMap map[schema.RefKey]map[schema.RefKey]bool) *consolepb.Config {
	return &consolepb.Config{
		//nolint:forcetypeassert
		Config:     decl.ToProto(),
		References: getReferencesFromMap(refMap, module, decl.Name),
	}
}

func dataFromDecl(decl *schema.Data, module string, refMap map[schema.RefKey]map[schema.RefKey]bool) *consolepb.Data {
	d := decl.ToProto()
	return &consolepb.Data{
		Data:       d,
		Schema:     schema.DataFromProto(d).String(),
		References: getReferencesFromMap(refMap, module, decl.Name),
	}
}

func databaseFromDecl(decl *schema.Database, module string, refMap map[schema.RefKey]map[schema.RefKey]bool) *consolepb.Database {
	return &consolepb.Database{
		Database:   decl.ToProto(),
		References: getReferencesFromMap(refMap, module, decl.Name),
	}
}

func enumFromDecl(decl *schema.Enum, module string, refMap map[schema.RefKey]map[schema.RefKey]bool) *consolepb.Enum {
	return &consolepb.Enum{
		Enum:       decl.ToProto(),
		References: getReferencesFromMap(refMap, module, decl.Name),
	}
}

func topicFromDecl(decl *schema.Topic, module string, refMap map[schema.RefKey]map[schema.RefKey]bool) *consolepb.Topic {
	return &consolepb.Topic{
		Topic:      decl.ToProto(),
		References: getReferencesFromMap(refMap, module, decl.Name),
	}
}

func typealiasFromDecl(decl *schema.TypeAlias, module string, refMap map[schema.RefKey]map[schema.RefKey]bool) *consolepb.TypeAlias {
	return &consolepb.TypeAlias{
		Typealias:  decl.ToProto(),
		References: getReferencesFromMap(refMap, module, decl.Name),
	}
}

func secretFromDecl(decl *schema.Secret, module string, refMap map[schema.RefKey]map[schema.RefKey]bool) *consolepb.Secret {
	return &consolepb.Secret{
		Secret:     decl.ToProto(),
		References: getReferencesFromMap(refMap, module, decl.Name),
	}
}

func verbFromDecl(decl *schema.Verb, sch *schema.Schema, module string, refMap map[schema.RefKey]map[schema.RefKey]bool) (*consolepb.Verb, error) {
	v := decl.ToProto()
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
	return &consolepb.Verb{
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

func (s *service) StreamModules(ctx context.Context, req *connect.Request[consolepb.StreamModulesRequest], stream *connect.ServerStream[consolepb.StreamModulesResponse]) error {

	err := s.sendStreamModulesResp(ctx, stream)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-s.schemaEventSource.Events():
			err = s.sendStreamModulesResp(ctx, stream)
			if err != nil {
				return err
			}
		}
	}
}

// filterDeployments removes any duplicate modules by selecting the deployment with the
// latest CreatedAt.
func (s *service) filterDeployments(unfilteredDeployments *schema.Schema) []*schema.Module {
	latest := make(map[string]*schema.Module)

	for _, deployment := range unfilteredDeployments.Modules {
		if existing, found := latest[deployment.Name]; !found || deployment.Runtime.Base.CreateTime.After(existing.Runtime.Base.CreateTime) {
			latest[deployment.Name] = deployment

		}
	}

	var result []*schema.Module
	for _, value := range latest {
		result = append(result, value)
	}

	return result
}

func (s *service) sendStreamModulesResp(ctx context.Context, stream *connect.ServerStream[consolepb.StreamModulesResponse]) error {
	unfilteredDeployments := s.schemaEventSource.View()

	deployments := s.filterDeployments(unfilteredDeployments)
	sch := &schema.Schema{
		Modules: deployments,
	}
	builtin := schema.Builtins()
	sch.Modules = append(sch.Modules, builtin)

	// Get topology
	sorted, err := buildengine.TopologicalSort(graph(sch))
	if err != nil {
		return fmt.Errorf("failed to sort modules: %w", err)
	}
	topology := &consolepb.Topology{
		Levels: make([]*consolepb.TopologyGroup, len(sorted)),
	}
	for i, level := range sorted {
		group := &consolepb.TopologyGroup{
			Modules: level,
		}
		topology.Levels[i] = group
	}

	refMap, err := getSchemaRefs(sch)
	if err != nil {
		return fmt.Errorf("failed to find references: %w", err)
	}

	var modules []*consolepb.Module
	for _, deployment := range deployments {
		if deployment.Runtime == nil || deployment.Runtime.Deployment == nil {
			continue
		}
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

	err = stream.Send(&consolepb.StreamModulesResponse{
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

func (s *service) GetConfig(ctx context.Context, req *connect.Request[consolepb.GetConfigRequest]) (*connect.Response[consolepb.GetConfigResponse], error) {
	resp, err := s.adminClient.ConfigGet(ctx, connect.NewRequest(&ftlv1.ConfigGetRequest{
		Ref: &ftlv1.ConfigRef{
			Module: req.Msg.Module,
			Name:   req.Msg.Name,
		},
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	return connect.NewResponse(&consolepb.GetConfigResponse{
		Value: resp.Msg.Value,
	}), nil
}

func (s *service) SetConfig(ctx context.Context, req *connect.Request[consolepb.SetConfigRequest]) (*connect.Response[consolepb.SetConfigResponse], error) {
	_, err := s.adminClient.ConfigSet(ctx, connect.NewRequest(&ftlv1.ConfigSetRequest{
		Ref: &ftlv1.ConfigRef{
			Module: req.Msg.Module,
			Name:   req.Msg.Name,
		},
		Value: req.Msg.Value,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to set config: %w", err)
	}
	return connect.NewResponse(&consolepb.SetConfigResponse{}), nil
}

func (s *service) GetSecret(ctx context.Context, req *connect.Request[consolepb.GetSecretRequest]) (*connect.Response[consolepb.GetSecretResponse], error) {
	resp, err := s.adminClient.SecretGet(ctx, connect.NewRequest(&ftlv1.SecretGetRequest{
		Ref: &ftlv1.ConfigRef{
			Name:   req.Msg.Name,
			Module: req.Msg.Module,
		},
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}
	return connect.NewResponse(&consolepb.GetSecretResponse{
		Value: resp.Msg.Value,
	}), nil
}

func (s *service) SetSecret(ctx context.Context, req *connect.Request[consolepb.SetSecretRequest]) (*connect.Response[consolepb.SetSecretResponse], error) {
	_, err := s.adminClient.SecretSet(ctx, connect.NewRequest(&ftlv1.SecretSetRequest{
		Ref: &ftlv1.ConfigRef{
			Name:   req.Msg.Name,
			Module: req.Msg.Module,
		},
		Value: req.Msg.Value,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to set secret: %w", err)
	}

	return connect.NewResponse(&consolepb.SetSecretResponse{}), nil
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

func (s *service) GetTimeline(ctx context.Context, req *connect.Request[timelinepb.GetTimelineRequest]) (*connect.Response[timelinepb.GetTimelineResponse], error) {
	resp, err := s.timelineClient.GetTimeline(ctx, connect.NewRequest(req.Msg))
	if err != nil {
		return nil, fmt.Errorf("failed to get timeline from service: %w", err)
	}
	return connect.NewResponse(resp.Msg), nil
}

func (s *service) StreamTimeline(ctx context.Context, req *connect.Request[timelinepb.StreamTimelineRequest], out *connect.ServerStream[timelinepb.StreamTimelineResponse]) error {
	stream, err := s.timelineClient.StreamTimeline(ctx, connect.NewRequest(req.Msg))
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

func (s *service) Status(ctx context.Context, req *connect.Request[ftlv1.StatusRequest]) (*connect.Response[ftlv1.StatusResponse], error) {
	resp, err := s.controllerClient.Status(ctx, connect.NewRequest(req.Msg))
	if err != nil {
		return nil, fmt.Errorf("failed to get status from controller: %w", err)
	}
	return connect.NewResponse(resp.Msg), nil
}

func (s *service) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	resp, err := s.callClient.Call(ctx, connect.NewRequest(req.Msg))
	if err != nil {
		return nil, fmt.Errorf("failed to call verb: %w", err)
	}
	return connect.NewResponse(resp.Msg), nil
}
