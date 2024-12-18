//go:build integration

package pubsub

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/IBM/sarama"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/exec"
	in "github.com/block/ftl/internal/integration"
	"github.com/block/ftl/internal/model"
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

		// After a deployment is "ready" it can take a second before a consumer group claims partitions.
		// "publisher.local" has "from=latest" so we need that group to be ready before we start publishing
		// otherwise it will start from the latest offset after claiming partitions.
		in.Sleep(time.Second*1),

		// publish half the events before subscriber is deployed
		publishToTestAndLocalTopics(calls/2),

		in.Deploy("subscriber"),

		// publish the other half of the events after subscriber is deployed
		publishToTestAndLocalTopics(calls/2),

		in.Sleep(time.Second*4),

		// check that there are the right amount of consumed events, depending on "from" offset option
		checkConsumed("publisher", "local", true, events, optional.None[string]()),
		checkConsumed("subscriber", "consume", true, events, optional.None[string]()),
		checkConsumed("subscriber", "consumeFromLatest", true, events/2, optional.None[string]()),
	)
}

func TestRetry(t *testing.T) {
	retriesPerCall := 2
	in.Run(t,
		in.WithLanguages("java", "go"),
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
	var deploymentKilledTime *time.Time
	in.Run(t,
		in.WithLanguages("go"),
		in.WithPubSub(),
		in.CopyModule("publisher"),
		in.CopyModule("subscriber"),
		in.Deploy("publisher"),
		in.Deploy("subscriber"),

		// consumer group must now have a member for each partition
		checkGroupMembership("subscriber", "consumeSlow", 1),

		// publish events that will take a long time to process on the first subscriber deployment
		// to test that rebalancing doesnt cause consumption to fail and skip events
		in.Repeat(100, in.Call("publisher", "publishSlow", in.Obj{}, func(t testing.TB, resp in.Obj) {})),

		// Upgrade deployment
		func(t testing.TB, ic in.TestContext) {
			in.Infof("Modifying code")
			path := filepath.Join(ic.WorkingDir(), "subscriber", "subscriber.go")

			bytes, err := os.ReadFile(path)
			assert.NoError(t, err)
			output := strings.ReplaceAll(string(bytes), "This deployment is TheFirstDeployment", "This deployment is TheSecondDeployment")
			assert.NoError(t, os.WriteFile(path, []byte(output), 0644))
		},
		in.Deploy("subscriber"),

		// Currently old deployment runs for a little bit longer.
		// During this time we expect the consumer group to have 2 members (old deployment and new deployment).
		// This will probably change when we have proper draining of the old deployment.
		checkGroupMembership("subscriber", "consumeSlow", 2),
		func(t testing.TB, ic in.TestContext) {
			in.Infof("Waiting for old deployment to be killed")
			start := time.Now()
			for {
				assert.True(t, time.Since(start) < 15*time.Second)
				ps, err := exec.Capture(ic.Context, ".", "ftl", "ps")
				assert.NoError(t, err)
				if strings.Count(string(ps), "dpl-subscriber-") == 1 {
					// original deployment has ended
					now := time.Now()
					deploymentKilledTime = &now
					return
				}
			}
		},
		// Once old deployment has ended, the consumer group should only have 1 member per partition (the new deployment)
		// This should happen fairly quickly. If it takes a while it could be because the previous deployment did not close
		// the group properly.
		checkGroupMembership("subscriber", "consumeSlow", 1),
		func(t testing.TB, ic in.TestContext) {
			assert.True(t, time.Since(*deploymentKilledTime) < 3*time.Second, "make sure old deployment was removed from consumer group fast enough")
		},

		// confirm that each message was consumed successfully
		checkConsumed("subscriber", "consumeSlow", true, 100, optional.None[string]()),
	)
}

func publishToTestAndLocalTopics(calls int) in.Action {
	// do this in parallel because we want to test race conditions
	return func(t testing.TB, ic in.TestContext) {
		actions := []in.Action{
			in.Repeat(calls, in.Call("publisher", "publishTen", in.Obj{}, func(t testing.TB, resp in.Obj) {})),
			in.Repeat(calls, in.Call("publisher", "publishTenLocal", in.Obj{}, func(t testing.TB, resp in.Obj) {})),
		}
		wg := &sync.WaitGroup{}
		for _, action := range actions {
			wg.Add(1)
			go func() {
				action(t, ic)
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func checkConsumed(module, verb string, success bool, count int, needle optional.Option[string]) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		if needle, ok := needle.Get(); ok {
			in.Infof("Checking for %v call(s) to %s.%s with needle %v", count, module, verb, needle)
		} else {
			in.Infof("Checking for %v call(s) to %s.%s", count, module, verb)
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
func checkGroupMembership(module, subscription string, expectedCount int) in.Action {
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
		assert.Equal(t, len(groups[0].Members), expectedCount, "expected consumer group %v to have %v members", consumerGroup, expectedCount)
	}
}
