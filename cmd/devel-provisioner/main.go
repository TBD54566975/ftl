package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/kong"
	"github.com/jpillora/backoff"

	provisioner "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1/provisionerconnect"
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
		plugin.WithEnvars("FTL_PROVISIONER_CF_DB_SECURITY_GROUP=sg-08e06d6f8327024de"),
	)
	if err != nil {
		panic(err)
	}

	desired := []*provisioner.Resource{{
		ResourceId: "foobardb",
		Resource: &provisioner.Resource_Postgres{
			Postgres: &provisioner.PostgresResource{},
		},
	}}

	req := &provisioner.ProvisionRequest{
		FtlClusterId:      "ftl-test-1",
		Module:            "test-module",
		ExistingResources: []*provisioner.Resource{},
		DesiredResources:  inContext(desired),
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

	retry := backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    5 * time.Second,
		Factor: 2,
	}
	for {
		println("polling: " + resp.Msg.ProvisioningToken)
		status, err := client.Client.Status(ctx, connect.NewRequest(&provisioner.StatusRequest{
			ProvisioningToken: resp.Msg.ProvisioningToken,
			DesiredResources:  desired,
		}))
		if err != nil {
			panic(err)
		}
		if success, ok := status.Msg.Status.(*provisioner.StatusResponse_Success); ok {
			println("finished!")
			for _, r := range success.Success.UpdatedResources {
				jsn, err := json.MarshalIndent(r, "", "  ")
				if err != nil {
					panic(err)
				}
				println(string(jsn))
			}
			break
		}
		time.Sleep(retry.Duration())
	}
	println("done")
}

func inContext(resources []*provisioner.Resource) []*provisioner.ResourceContext {
	var result []*provisioner.ResourceContext
	for _, r := range resources {
		result = append(result, &provisioner.ResourceContext{Resource: r})
	}
	return result
}
