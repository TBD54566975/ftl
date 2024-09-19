package main

import (
	"context"
	"os"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/kong"
	"github.com/jpillora/backoff"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/internal/log"
)

var cli struct {
	log.Config `prefix:"log-"`
}

// For locally testing the provisioners.
// You should add the provisioner being tested to your PATH before running this
func main() {
	kong.Parse(&cli)
	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.Config))

	client, ctx, err := plugin.Spawn(
		ctx,
		log.Debug,
		"ftl-provisioner-cloudformation",
		".",
		"ftl-provisioner-cloudformation",
		provisionerconnect.NewProvisionerPluginServiceClient,
		plugin.WithEnvars("FTL_PROVISIONER_CF_DB_SUBNET_GROUP=aurora-postgres-subnet-group"),
	)
	if err != nil {
		panic(err)
	}

	req := &provisioner.ProvisionRequest{
		FtlClusterId:      "ftl-test-1",
		Module:            "test-module",
		ExistingResources: []*provisioner.Resource{},
		DesiredResources: []*provisioner.ResourceContext{{
			Resource: &provisioner.Resource{
				ResourceId: "foodb",
				Resource: &provisioner.Resource_Postgres{
					Postgres: &provisioner.PostgresResource{},
				},
			},
		}},
	}

	plan, err := client.Client.Plan(ctx, connect.NewRequest(&provisioner.PlanRequest{
		Provisioning: req,
	}))
	if err != nil {
		panic(err)
	}
	println("### PLAN ###")
	println(plan.Msg.Plan)

	println("### EXECUTION ###")
	resp, err := client.Client.Provision(ctx, connect.NewRequest(req))
	if err != nil {
		panic(err)
	}
	if resp.Msg.Status == provisioner.ProvisionResponse_NO_CHANGES {
		println("no changes")
		return
	}

	retry := backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    5 * time.Second,
		Factor: 2,
	}
	for {
		println("polling: " + resp.Msg.ProvisioningToken)
		status, err := client.Client.Status(ctx, connect.NewRequest(&provisioner.StatusRequest{
			ProvisioningToken: resp.Msg.ProvisioningToken,
		}))
		if err != nil {
			panic(err)
		}
		if fail, ok := status.Msg.Status.(*provisioner.StatusResponse_Failed); ok {
			panic(fail.Failed.ErrorMessage)
		} else if success, ok := status.Msg.Status.(*provisioner.StatusResponse_Success); ok {
			println("finished!")
			for _, p := range success.Success.Properties {
				println("  ", p.ResourceId, "\t", p.Key, "\t", p.Value)
			}
			break
		}
		time.Sleep(retry.Duration())
	}
	println("done")
}
