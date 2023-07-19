package controller

import (
	"context"
	"github.com/TBD54566975/ftl/internal/log"
	"time"

	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/common/model"
	"github.com/TBD54566975/ftl/controller/internal/dal"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/schema"
)

type Call struct {
	requestKey    model.IngressRequestKey
	runnerKey     model.RunnerKey
	controllerKey model.ControllerKey
	startTime     time.Time
	destVerb      *schema.VerbRef
	callers       []*schema.VerbRef
	request       *ftlv1.CallRequest
	response      *ftlv1.CallResponse
}

func (s *Service) recordCall(ctx context.Context, call *Call) error {
	sourceVerb := schema.VerbRef{}
	if len(call.callers) > 0 {
		sourceVerb = *call.callers[0]
	}

	var responseError error
	if call.response.GetError() != nil {
		responseError = errors.New(call.response.GetError().GetMessage())
	}
	err := s.dal.InsertCallEntry(ctx, &dal.CallEntry{
		RequestKey:    call.requestKey,
		RunnerKey:     call.runnerKey,
		ControllerKey: call.controllerKey,
		Duration:      time.Since(call.startTime),
		SourceVerb:    sourceVerb,
		DestVerb:      *call.destVerb,
		Request:       call.request.GetBody(),
		Response:      call.response.GetBody(),
		Error:         responseError,
	})

	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// recordCallError records a call that failed to be made.
//
// This is used when a call fails, but we still want to record the error.
// Because of that, we only log failures within this function vs. returning another error.
func (s *Service) recordCallError(ctx context.Context, call *call, callError error) {
	logger := log.FromContext(ctx)
	sourceVerb := schema.VerbRef{}
	if len(call.callers) > 0 {
		sourceVerb = *call.callers[0]
	}

	err := s.dal.InsertCallEntry(ctx, &dal.CallEntry{
		RequestKey:    call.requestKey,
		RunnerKey:     call.runnerKey,
		ControllerKey: call.controllerKey,
		Duration:      time.Since(call.startTime),
		SourceVerb:    sourceVerb,
		DestVerb:      *call.destVerb,
		Request:       call.request.GetBody(),
		Response:      call.response.GetBody(),
		Error:         callError,
	})

	if err != nil {
		logger.Errorf(err, "failed to record call error")
	}
}
