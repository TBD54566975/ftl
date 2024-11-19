package terminal

import (
	"context"

	"connectrpc.com/connect"
	"github.com/posener/complete"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
)

func Predictors(ctx context.Context, client ftlv1connect.SchemaServiceClient) map[string]complete.Predictor {
	return map[string]complete.Predictor{
		"verbs": &verbPredictor{
			Client: client,
			Ctx:    ctx,
		},
	}
}

type verbPredictor struct {
	Client ftlv1connect.SchemaServiceClient
	Ctx    context.Context
}

func (v *verbPredictor) Predict(args complete.Args) []string {
	response, err := v.Client.GetSchema(v.Ctx, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
	if err != nil {
		// Do we want to report errors here?
		return nil
	}
	ret := []string{}
	for _, module := range response.Msg.Schema.Modules {
		for _, dec := range module.Decls {
			if dec.GetVerb() != nil {
				verb := module.Name + "." + dec.GetVerb().Name
				ret = append(ret, verb)
			}
		}
	}
	return ret
}
