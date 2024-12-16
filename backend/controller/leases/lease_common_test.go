//go:build infrastructure || integration

package leases_test

import (
	"fmt"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	"golang.org/x/sync/errgroup"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	in "github.com/block/ftl/internal/integration"
)

func setupLeaseTests() []in.ActionOrOption {
	return mapValues(
		in.WithLanguages("go", "java"),
		in.CopyModule("leases"),
		in.Build("leases"),
		// checks if leases work in a unit test environment
		in.IfLanguage("go", in.ExecModuleTest("leases")),
		in.Deploy("leases"),
		// checks if it leases work with a real controller
		func(t testing.TB, ic in.TestContext) {
			// Start a lease.
			wg := errgroup.Group{}
			wg.Go(func() error {
				in.Infof("Acquiring lease")
				resp, err := ic.Verbs.Call(ic, connect.NewRequest(&ftlv1.CallRequest{
					Verb: &schemapb.Ref{Module: "leases", Name: "acquire"},
					Body: []byte("{}"),
				}))
				if err != nil {
					return err
				}
				assert.NoError(t, err)
				if respErr := resp.Msg.GetError(); respErr != nil {
					return fmt.Errorf("received error on first call: %v", respErr)
				}
				return err
			})

			time.Sleep(time.Second)

			in.Infof("Trying to acquire lease again")
			// Trying to obtain the lease again should fail.
			resp, err := ic.Verbs.Call(ic, connect.NewRequest(&ftlv1.CallRequest{
				Verb: &schemapb.Ref{Module: "leases", Name: "acquire"},
				Body: []byte("{}"),
			}))
			assert.NoError(t, err)
			if resp.Msg.GetError() == nil {
				t.Fatalf("expected error but got: %#v", resp.Msg.GetError())
			}
			err = wg.Wait()
			assert.NoError(t, err)
		},
	)
}

func mapValues(option ...in.ActionOrOption) []in.ActionOrOption {
	return option
}
