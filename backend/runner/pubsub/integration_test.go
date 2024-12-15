//go:build integration

package pubsub

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/IBM/sarama"
	timelinepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/TBD54566975/ftl/common/slices"
	in "github.com/TBD54566975/ftl/internal/integration"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
)

func TestPubSub(t *testing.T) {
	calls := 20
	events := calls * 10
	in.Run(t,
		in.WithLanguages("java", "go", "kotlin"),
		in.WithPubSub(),
		in.CopyModule("publisher"),
		in.CopyModule("subscriber"),
		in.Deploy("publisher"),
		in.Deploy("subscriber"),

		// publish events
		in.Repeat(calls, in.Call("publisher", "publishTen", in.Obj{}, func(t testing.TB, resp in.Obj) {})),

		in.Sleep(time.Second*4),

		// check that there are the right amount of successful async calls
		checkConsumed("subscriber", "consume", true, events, optional.None[string]()),
	)
}

func TestRetry(t *testing.T) {
	retriesPerCall := 2
	in.Run(t,
		in.WithLanguages("java", "go"),
		in.WithPubSub(),
		in.CopyModule("publisher"),
		in.CopyModule("subscriber"),
		in.Deploy("publisher"),
		in.Deploy("subscriber"),

		// publish events
		in.Call("publisher", "publishOneToTopic2", map[string]any{"haystack": "firstCall"}, func(t testing.TB, resp in.Obj) {}),
		in.Call("publisher", "publishOneToTopic2", map[string]any{"haystack": "secondCall"}, func(t testing.TB, resp in.Obj) {}),

		in.Sleep(time.Second*7),

		checkConsumed("subscriber", "consumeButFailAndRetry", false, retriesPerCall+1, optional.Some("firstCall")),
		checkConsumed("subscriber", "consumeButFailAndRetry", false, retriesPerCall+1, optional.Some("secondCall")),
	)
}

func TestExternalPublishRuntimeCheck(t *testing.T) {
	// No java as there is no API for this
	in.Run(t,
		in.CopyModule("publisher"),
		in.CopyModule("subscriber"),
		in.Deploy("publisher"),
		in.Deploy("subscriber"),

		in.ExpectError(
			in.Call("subscriber", "publishToExternalModule", in.Obj{}, func(t testing.TB, resp in.Obj) {}),
			"can not publish to another module's topic",
		),
	)
}

// TestConsumerGroupMembership tests that when a runner ends, the consumer group is properly exited.
func TestConsumerGroupMembership(t *testing.T) {
	in.Run(t,
		in.WithLanguages("java", "go"),
		in.WithPubSub(),
		in.CopyModule("publisher"),
		in.CopyModule("subscriber"),
		in.Deploy("publisher"),
		in.Deploy("subscriber"),

		// consumer group must now have a member
		checkGroupMembership("subscriber", "consume", true),

		// Stop subscriber deployment
		func(t testing.TB, ic in.TestContext) {
			in.ExecWithOutput("ftl", []string{"ps", "--json"}, func(jsonStr string) {
				// parse newline delimted json
				decoder := json.NewDecoder(strings.NewReader(jsonStr))
				for decoder.More() {
					var deployment map[string]any
					if err := decoder.Decode(&deployment); err != nil {
						assert.NoError(t, err)
					}
					depName, ok := deployment["deployment"].(string)
					assert.True(t, ok)
					if strings.Contains(depName, "subscriber") {
						in.Exec("ftl", "kill", depName)(t, ic)
						return
					}
				}
				assert.True(t, false, "subscriber deployment not found")
			})(t, ic)
		},

		in.Sleep(time.Second*2),

		in.WithoutRetries(checkGroupMembership("subscriber", "consume", false)),
	)
}

func checkConsumed(module, verb string, success bool, count int, needle optional.Option[string]) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		if needle, ok := needle.Get(); ok {
			in.Infof("Checking for %v call(s) with needle %v", count, needle)
		} else {
			in.Infof("Checking for %v call(s)", count)
		}
		resp, err := ic.Timeline.GetTimeline(ic.Context, connect.NewRequest(&timelinepb.GetTimelineRequest{
			Limit: 100000,
			Filters: []*timelinepb.GetTimelineRequest_Filter{
				{
					Filter: &timelinepb.GetTimelineRequest_Filter_EventTypes{
						EventTypes: &timelinepb.GetTimelineRequest_EventTypeFilter{
							EventTypes: []timelinepb.EventType{
								timelinepb.EventType_EVENT_TYPE_CALL,
							},
						},
					},
				},
				{
					Filter: &timelinepb.GetTimelineRequest_Filter_Module{
						Module: &timelinepb.GetTimelineRequest_ModuleFilter{
							Module: module,
							Verb:   &verb,
						},
					},
				},
			},
		}))
		assert.NoError(t, err)
		calls := slices.Filter(slices.Map(resp.Msg.Events, func(e *timelinepb.Event) *timelinepb.CallEvent {
			return e.GetCall()
		}), func(c *timelinepb.CallEvent) bool {
			if c == nil {
				return false
			}
			assert.NotEqual(t, nil, c.RequestKey, "pub sub calls need a request key")
			requestKey, err := model.ParseRequestKey(*c.RequestKey)
			assert.NoError(t, err)
			assert.Equal(t, requestKey.Payload.Origin, model.OriginPubsub, "expected pubsub origin")
			if needle, ok := needle.Get(); ok && !strings.Contains(c.Request, needle) {
				return false
			}
			return true
		})
		successfulCalls := slices.Filter(calls, func(call *timelinepb.CallEvent) bool {
			return call.Error == nil
		})
		unsuccessfulCalls := slices.Filter(calls, func(call *timelinepb.CallEvent) bool {
			return call.Error != nil
		})
		if success {
			assert.Equal(t, count, len(successfulCalls), "expected %v successful calls (failed calls: %v)", count, len(unsuccessfulCalls))
		} else {
			assert.Equal(t, count, len(unsuccessfulCalls), "expected %v unsuccessful calls (successful calls: %v)", count, len(successfulCalls))
		}
	}
}

func checkGroupMembership(module, subscription string, expected bool) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		consumerGroup := module + "." + subscription
		in.Infof("Checking group membership for %v", consumerGroup)

		client, err := sarama.NewClient(in.RedPandaBrokers, sarama.NewConfig())
		assert.NoError(t, err)
		defer client.Close()

		clusterAdmin, err := sarama.NewClusterAdminFromClient(client)
		assert.NoError(t, err)
		defer clusterAdmin.Close()

		groups, err := clusterAdmin.DescribeConsumerGroups([]string{consumerGroup})
		assert.NoError(t, err)
		assert.Equal(t, len(groups), 1)
		if expected {
			assert.True(t, len(groups[0].Members) > 0, "expected consumer group %v to have members", consumerGroup)
		} else {
			assert.False(t, len(groups[0].Members) > 0, "expected consumer group %v to have no members", consumerGroup)
		}
	}
}
