package controller

import (
	"context"
	"github.com/TBD54566975/ftl/common/model"
	"github.com/TBD54566975/ftl/controller/internal/dal"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/schema"
	"github.com/alecthomas/errors"
	"time"
)

type call struct {
	requestID     int64
	runnerKey     model.RunnerKey
	controllerKey model.ControllerKey
	startTime     time.Time
	destVerb      *schema.VerbRef
	callers       []*schema.VerbRef
	request       *ftlv1.CallRequest
	response      *ftlv1.CallResponse
}

func (s *Service) recordCall(ctx context.Context, call *call) error {
	sourceVerb := schema.VerbRef{}
	if len(call.callers) > 0 {
		sourceVerb = *call.callers[0]
	}

	var callError error
	if call.response.GetError() != nil {
		callError = errors.New(call.response.GetError().GetMessage())
	}
	err := s.dal.InsertCallEntry(ctx, &dal.Call{
		RequestID:     call.requestID,
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
		return errors.WithStack(err)
	}
	return nil
}
