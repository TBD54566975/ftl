//go:build infrastructure

package scaling_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/atomic"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	in "github.com/TBD54566975/ftl/internal/integration"
)

func TestKubeScaling(t *testing.T) {
	failure := atomic.Value[error]{}
	done := atomic.Value[bool]{}
	done.Store(false)
	routineStopped := sync.WaitGroup{}
	routineStopped.Add(1)
	echoDeployment := map[string]string{}
	in.Run(t,
		in.WithKubernetes(),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.CopyModule("naughty"),
		in.Deploy("naughty"),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Hello, Bob!!!", response)
		}),
		in.VerifyKubeState(func(ctx context.Context, t testing.TB, namespace string, client kubernetes.Clientset) {
			deps, err := client.AppsV1().Deployments(namespace).List(ctx, v1.ListOptions{})
			assert.NoError(t, err)
			for _, dep := range deps.Items {
				if strings.HasPrefix(dep.Name, "dpl-echo") {
					echoDeployment["name"] = dep.Name
				}
			}
			assert.NotEqual(t, "", echoDeployment["name"])
		}),
		in.Call("naughty", "beNaughty", echoDeployment, func(t testing.TB, response string) {
			// If istio is not present we should be able to ping the echo service directly.
			// Istio should prevent this
			assert.Equal(t, strconv.FormatBool(false), response)
		}),
		func(t testing.TB, ic in.TestContext) {
			// Hit the verb constantly to test rolling updates.
			go func() {
				defer func() {
					if r := recover(); r != nil {
						failure.Store(fmt.Errorf("panic in verb: %v at %v", r, time.Now()))
					}
					routineStopped.Done()
				}()
				for !done.Load() {
					in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
						if !strings.Contains(response, "Bob") {
							failure.Store(fmt.Errorf("unexpected response: %s", response))
							return
						}
					})(t, ic)
				}
			}()
		},
		in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "Hello", "Bye"))
		}, "echo.go"),
		in.Deploy("echo"),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Bye, Bob!!!", response)
		}),
		func(t testing.TB, ic in.TestContext) {
			err := failure.Load()
			assert.NoError(t, err)
		},
		in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "Bye", "Bonjour"))
		}, "echo.go"),
		in.Deploy("echo"),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Bonjour, Bob!!!", response)
		}),
		func(t testing.TB, ic in.TestContext) {
			done.Store(true)
			routineStopped.Wait()
			err := failure.Load()
			assert.NoError(t, err)
		},
		in.VerifyKubeState(func(ctx context.Context, t testing.TB, namespace string, client kubernetes.Clientset) {
			deps, err := client.AppsV1().Deployments(namespace).List(ctx, v1.ListOptions{})
			assert.NoError(t, err)
			depCount := 0
			for _, dep := range deps.Items {
				if strings.HasPrefix(dep.Name, "dpl-echo") {
					depCount++
					service, err := client.CoreV1().Services(namespace).Get(ctx, dep.Name, v1.GetOptions{})
					assert.NoError(t, err)
					assert.Equal(t, 1, len(dep.OwnerReferences))
					assert.Equal(t, service.UID, dep.OwnerReferences[0].UID)
				}
			}
			assert.Equal(t, 1, depCount)
		}),
	)
}
