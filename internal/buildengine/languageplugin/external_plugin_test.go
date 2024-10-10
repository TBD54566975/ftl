package languageplugin

import (
	"context"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/atomic"
	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/either"
	"github.com/alecthomas/types/optional"

	"connectrpc.com/connect"
	langpb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language"
	"github.com/TBD54566975/ftl/internal/builderrors"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/schema"
)

type testBuildContext struct {
	BuildContext
	ContextId string
	IsRebuild bool
}

type mockExternalPluginClient struct {
	flags []*langpb.GetCreateModuleFlagsResponse_Flag

	buildEvents        chan *langpb.BuildEvent
	latestBuildContext atomic.Value[testBuildContext]
}

var _ externalPluginClient = &mockExternalPluginClient{}

func newMockExternalPlugin() *mockExternalPluginClient {
	return &mockExternalPluginClient{
		buildEvents: make(chan *langpb.BuildEvent, 64),
	}
}

func (p *mockExternalPluginClient) getCreateModuleFlags(context.Context, *connect.Request[langpb.GetCreateModuleFlagsRequest]) (*connect.Response[langpb.GetCreateModuleFlagsResponse], error) {
	return connect.NewResponse(&langpb.GetCreateModuleFlagsResponse{
		Flags: p.flags,
	}), nil
}

func (p *mockExternalPluginClient) createModule(context.Context, *connect.Request[langpb.CreateModuleRequest]) (*connect.Response[langpb.CreateModuleResponse], error) {
	panic("not implemented")
}

func (p *mockExternalPluginClient) moduleConfigDefaults(ctx context.Context, req *connect.Request[langpb.ModuleConfigDefaultsRequest]) (*connect.Response[langpb.ModuleConfigDefaultsResponse], error) {
	return connect.NewResponse(&langpb.ModuleConfigDefaultsResponse{
		DeployDir: "test-deploy-dir",
		Watch:     []string{"a", "b", "c"},
	}), nil
}

func (p *mockExternalPluginClient) getDependencies(context.Context, *connect.Request[langpb.DependenciesRequest]) (*connect.Response[langpb.DependenciesResponse], error) {
	panic("not implemented")
}

func buildContextFromProto(proto *langpb.BuildContext) (BuildContext, error) {
	sch, err := schema.FromProto(proto.Schema)
	if err != nil {
		return BuildContext{}, err
	}
	return BuildContext{
		Schema:       sch,
		Dependencies: proto.Dependencies,
		Config: moduleconfig.ModuleConfig{
			Dir:                proto.ModuleConfig.Path,
			Language:           "test",
			Realm:              "test",
			Module:             proto.ModuleConfig.Name,
			Build:              optional.Ptr(proto.ModuleConfig.Build).Default(""),
			DeployDir:          proto.ModuleConfig.DeployDir,
			GeneratedSchemaDir: optional.Ptr(proto.ModuleConfig.GeneratedSchemaDir).Default(""),
			Watch:              proto.ModuleConfig.Watch,
			LanguageConfig:     proto.ModuleConfig.LanguageConfig.AsMap(),
		},
	}, nil
}

func (p *mockExternalPluginClient) build(ctx context.Context, req *connect.Request[langpb.BuildRequest]) (chan *langpb.BuildEvent, streamCancelFunc, error) {
	bctx, err := buildContextFromProto(req.Msg.BuildContext)
	if err != nil {
		return nil, nil, err
	}
	p.latestBuildContext.Store(testBuildContext{
		BuildContext: bctx,
		ContextId:    req.Msg.BuildContext.Id,
		IsRebuild:    false,
	})
	return p.buildEvents, func() {}, nil
}

func (p *mockExternalPluginClient) buildContextUpdated(ctx context.Context, req *connect.Request[langpb.BuildContextUpdatedRequest]) (*connect.Response[langpb.BuildContextUpdatedResponse], error) {
	bctx, err := buildContextFromProto(req.Msg.BuildContext)
	if err != nil {
		return nil, err
	}
	p.latestBuildContext.Store(testBuildContext{
		BuildContext: bctx,
		ContextId:    req.Msg.BuildContext.Id,
		IsRebuild:    true,
	})
	return connect.NewResponse(&langpb.BuildContextUpdatedResponse{}), nil
}

func (p *mockExternalPluginClient) kill() error {
	return nil
}

func setUp() (context.Context, *externalPlugin, *mockExternalPluginClient, BuildContext) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	mockImpl := newMockExternalPlugin()
	plugin := newExternalPluginForTesting(ctx, mockImpl)

	bctx := BuildContext{
		Config: moduleconfig.ModuleConfig{
			Module:   "name",
			Dir:      "test/dir",
			Language: "test-lang",
		},
		Schema:       &schema.Schema{},
		Dependencies: []string{},
	}
	return ctx, plugin, mockImpl, bctx
}

func TestCreateModuleFlags(t *testing.T) {
	for _, tt := range []struct {
		protoFlags    []*langpb.GetCreateModuleFlagsResponse_Flag
		expectedFlags []*kong.Flag
		expectedError optional.Option[string]
	}{
		{
			protoFlags: []*langpb.GetCreateModuleFlagsResponse_Flag{
				{
					Name:        "full-flag",
					Help:        "This has all the fields set",
					Envar:       optional.Some("full-flag").Ptr(),
					Short:       optional.Some("f").Ptr(),
					Placeholder: optional.Some("placeholder").Ptr(),
					Default:     optional.Some("defaultValue").Ptr(),
				},
				{
					Name: "sparse-flag",
					Help: "This has only the minimum fields set",
				},
			},
			expectedFlags: []*kong.Flag{
				{
					Value: &kong.Value{
						Name:       "full-flag",
						Help:       "This has all the fields set",
						HasDefault: true,
						Default:    "defaultValue",
						Tag: &kong.Tag{
							Envs: []string{
								"full-flag",
							},
						},
					},
					PlaceHolder: "placeholder",
					Short:       'f',
				},
				{
					Value: &kong.Value{
						Name: "sparse-flag",
						Help: "This has only the minimum fields set",
						Tag:  &kong.Tag{},
					},
				},
			},
		},
		{
			protoFlags: []*langpb.GetCreateModuleFlagsResponse_Flag{
				{
					Name:  "multi-char-short",
					Help:  "This has all the fields set",
					Short: optional.Some("multi").Ptr(),
				},
			},
			expectedError: optional.Some(`invalid flag declared: short flag "multi" for multi-char-short must be a single character`),
		},
		{
			protoFlags: []*langpb.GetCreateModuleFlagsResponse_Flag{
				{
					Name:  "dupe-short-1",
					Help:  "Short must be unique",
					Short: optional.Some("d").Ptr(),
				},
				{
					Name:  "dupe-short-2",
					Help:  "Short must be unique",
					Short: optional.Some("d").Ptr(),
				},
			},
			expectedError: optional.Some(`multiple flags declared with the same short name: dupe-short-1 and dupe-short-2`),
		},
	} {
		t.Run(tt.protoFlags[0].Name, func(t *testing.T) {
			t.Parallel()

			ctx, plugin, mockImpl, _ := setUp()
			mockImpl.flags = tt.protoFlags
			kongFlags, err := plugin.GetCreateModuleFlags(ctx)
			if expectedError, ok := tt.expectedError.Get(); ok {
				assert.Contains(t, err.Error(), expectedError)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedFlags, kongFlags)
		})
	}
}

func TestSimultaneousBuild(t *testing.T) {
	t.Parallel()
	ctx, plugin, _, bctx := setUp()
	_ = beginBuild(ctx, plugin, bctx, false)
	r := beginBuild(ctx, plugin, bctx, false)
	result, ok := (<-r).(either.Right[BuildResult, error])
	assert.True(t, ok, "expected error, got %v", result)
	assert.Contains(t, result.Get().Error(), "build already in progress")
}

func TestMismatchedBuildContextId(t *testing.T) {
	t.Parallel()
	ctx, plugin, mockImpl, bctx := setUp()

	// build
	result := beginBuild(ctx, plugin, bctx, false)

	// send mismatched build result (ie: a different build attempt completing)
	mockImpl.buildEvents <- buildEventWithBuildError("fake", false, "this is not the result you are looking for")

	// send automatic rebuild result for the same context id (should be ignored)
	realId := mockImpl.latestBuildContext.Load().ContextId
	mockImpl.buildEvents <- buildEventWithBuildError(realId, true, "this is not the result you are looking for")

	// send real build result
	mockImpl.buildEvents <- buildEventWithBuildError(realId, false, "this is the correct result")

	// check result
	checkResult(t, <-result, "this is the correct result")
}

func TestRebuilds(t *testing.T) {
	t.Parallel()
	ctx, plugin, mockImpl, bctx := setUp()

	// build and activate automatic rebuilds
	result := beginBuild(ctx, plugin, bctx, true)

	// send first build result
	testBuildCtx := mockImpl.latestBuildContext.Load()
	mockImpl.buildEvents <- buildEventWithBuildError(testBuildCtx.ContextId, false, "first build")

	// check result
	checkResult(t, <-result, "first build")

	// send rebuild request with updated schema
	bctx.Schema.Modules = append(bctx.Schema.Modules, &schema.Module{Name: "another"})
	sch, err := schema.ValidateSchema(bctx.Schema)
	assert.NoError(t, err, "schema should be valid")
	result = beginBuild(ctx, plugin, bctx, true)

	// send rebuild result
	testBuildCtx = mockImpl.latestBuildContext.Load()
	assert.Equal(t, testBuildCtx.Schema, sch, "schema should have been updated")
	mockImpl.buildEvents <- buildEventWithBuildError(testBuildCtx.ContextId, false, "second build")

	// check rebuild result
	checkResult(t, <-result, "second build")
}

func TestAutomaticRebuilds(t *testing.T) {
	t.Parallel()
	ctx, plugin, mockImpl, bctx := setUp()

	updates := make(chan PluginEvent, 64)
	plugin.Updates().Subscribe(updates)

	// build and activate automatic rebuilds
	result := beginBuild(ctx, plugin, bctx, true)

	// plugin sends auto rebuild has started event (should be ignored)
	mockImpl.buildEvents <- &langpb.BuildEvent{
		Event: &langpb.BuildEvent_AutoRebuildStarted{},
	}
	// plugin sends auto rebuild event (should be ignored)
	mockImpl.buildEvents <- buildEventWithBuildError("fake", true, "auto rebuild to ignore")

	// send first build result
	time.Sleep(500 * time.Millisecond)
	buildCtx := mockImpl.latestBuildContext.Load()
	mockImpl.buildEvents <- buildEventWithBuildError(buildCtx.ContextId, false, "first build")

	// check result
	checkResult(t, <-result, "first build")

	// confirm that nothing was posted to Updates() (ie: the auto-rebuilds events were ignored)
	select {
	case <-updates:
		t.Fatalf("expected auto rebuilds events to not get published while build is in progress")
	case <-time.After(2 * time.Second):
		// as expected, no events published plugin
	}

	// plugin sends auto rebuild events
	mockImpl.buildEvents <- &langpb.BuildEvent{
		Event: &langpb.BuildEvent_AutoRebuildStarted{},
	}
	mockImpl.buildEvents <- buildEventWithBuildError(buildCtx.ContextId, true, "first real auto rebuild")
	// plugin sends auto rebuild events again (this time with no rebuild started event)
	mockImpl.buildEvents <- buildEventWithBuildError(buildCtx.ContextId, true, "second real auto rebuild")

	// confirm that auto rebuilds events were published
	events := eventsFromChannel(updates)
	assert.Equal(t, len(events), 3, "expected 3 events")
	assert.Equal(t, PluginEvent(AutoRebuildStartedEvent{Module: bctx.Config.Module}), events[0])
	checkAutoRebuildResult(t, events[1], "first real auto rebuild")
	checkAutoRebuildResult(t, events[2], "second real auto rebuild")
}

func eventsFromChannel(updates chan PluginEvent) []PluginEvent {
	// wait a bit to let events get published
	time.Sleep(200 * time.Millisecond)

	events := []PluginEvent{}
	for {
		select {
		case e := <-updates:
			events = append(events, e)
		default:
			// no more events available right now
			return events
		}
	}
}

func buildEventWithBuildError(contextId string, isAutomaticRebuild bool, msg string) *langpb.BuildEvent {
	return &langpb.BuildEvent{
		Event: &langpb.BuildEvent_BuildFailure{
			BuildFailure: &langpb.BuildFailure{
				ContextId:          contextId,
				IsAutomaticRebuild: isAutomaticRebuild,
				Errors: langpb.ErrorsToProto([]builderrors.Error{
					{
						Msg: msg,
					},
				}),
			},
		},
	}
}

func beginBuild(ctx context.Context, plugin *externalPlugin, bctx BuildContext, autoRebuild bool) chan either.Either[BuildResult, error] {
	result := make(chan either.Either[BuildResult, error])
	go func() {
		r, err := plugin.Build(ctx, "", bctx, []string{}, autoRebuild)
		if err != nil {
			result <- either.RightOf[BuildResult](err)
		} else {
			result <- either.LeftOf[error](r)
		}
	}()
	// sleep to make sure impl has received the build context
	time.Sleep(500 * time.Millisecond)
	return result
}

func checkResult(t *testing.T, r either.Either[BuildResult, error], expectedMsg string) {
	t.Helper()
	left, ok := r.(either.Left[BuildResult, error])
	assert.True(t, ok, "expected build result, got %v", r)

	buildResult := left.Get()
	assert.Equal(t, len(buildResult.Errors), 1)
	assert.Equal(t, buildResult.Errors[0].Msg, expectedMsg)
}

func checkAutoRebuildResult(t *testing.T, e PluginEvent, expectedMsg string) {
	t.Helper()
	event, ok := e.(AutoRebuildEndedEvent)
	assert.True(t, ok, "expected auto rebuild event, got %v", e)
	checkResult(t, event.Result, expectedMsg)
}
