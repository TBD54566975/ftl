//go:build infrastructure

package scaling_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

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
	in.Run(t,
		in.WithKubernetes(),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Hello, Bob!!!", response)
		}),
		in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "Hello", "Bye"))
		}, "echo.go"),
		func(t testing.TB, ic in.TestContext) {
			// Hit the verb constantly to test rolling updates.
			go func() {
				defer routineStopped.Done()
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
		in.Deploy("echo"),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Bye, Bob!!!", response)
		}),
		func(t testing.TB, ic in.TestContext) {
			done.Store(true)
			routineStopped.Wait()
			err := failure.Load()
			assert.NoError(t, err)
		},
		in.VerifyKubeState(func(ctx context.Context, t testing.TB, namespace string, client *kubernetes.Clientset) {
			deps, err := client.AppsV1().Deployments(namespace).List(ctx, v1.ListOptions{})
			assert.NoError(t, err)
			depCount := 0
			for _, dep := range deps.Items {
				if strings.HasPrefix(dep.Name, "dpl-echo") || strings.HasPrefix(dep.Name, "dpl-time") {
					depCount++
					service, err := client.CoreV1().Services(namespace).Get(ctx, dep.Name, v1.GetOptions{})
					assert.NoError(t, err)
					assert.Equal(t, 1, len(service.OwnerReferences))
					assert.Equal(t, dep.UID, service.OwnerReferences[0].UID)
				}
			}
			assert.Equal(t, 1, depCount)
		}),
	)
}