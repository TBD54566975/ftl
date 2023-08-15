package controller

import (
	"context"
	"time"

	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/model"
	"github.com/TBD54566975/ftl/backend/controller/internal/dal"
	"github.com/TBD54566975/ftl/backend/schema"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
)

type Call struct {
	deploymentKey model.DeploymentKey
	requestKey    model.IngressRequestKey
	startTime     time.Time
	destVerb      *schema.VerbRef
	callers       []*schema.VerbRef
	request       *ftlv1.CallRequest
	response      *ftlv1.CallResponse
}

// recordCallError records a call that failed to be made.
//
// This is used when a call fails, but we still want to record the error.
// Because of that, we only log failures within this function vs. returning another error.
func (s *Service) recordCall(ctx context.Context, call *Call) {
	var responseError error
	if e := call.response.GetError(); e != nil {
		responseError = errors.New(e.GetMessage())
	}
	s.recordCallError(ctx, call, responseError)
}

func (s *Service) recordCallError(ctx context.Context, call *Call, callError error) {
	logger := log.FromContext(ctx)
	sourceVerb := schema.VerbRef{}
	if len(call.callers) > 0 {
		sourceVerb = *call.callers[0]
	}

	err := s.dal.InsertCallEntry(ctx, &dal.CallEntry{
		DeploymentKey: call.deploymentKey,
		RequestKey:    call.requestKey,
		Duration:      time.Since(call.startTime),
		SourceVerb:    sourceVerb,
		DestVerb:      *call.destVerb,
		Request:       call.request.GetBody(),
		Response:      call.response.GetBody(),
		Error:         callError,
	})
	if err != nil {
		logger.Errorf(err, "failed to record call")
	}
}
