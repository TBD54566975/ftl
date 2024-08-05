package controller

import (
	"context"
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/dal"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
)

type Call struct {
	deploymentKey model.DeploymentKey
	requestKey    model.RequestKey
	startTime     time.Time
	destVerb      *schema.Ref
	callers       []*schema.Ref
	request       *ftlv1.CallRequest
	response      optional.Option[*ftlv1.CallResponse]
	callError     optional.Option[error]
}

func (s *Service) recordCall(ctx context.Context, call *Call) {
	logger := log.FromContext(ctx)
	var sourceVerb optional.Option[schema.Ref]
	if len(call.callers) > 0 {
		sourceVerb = optional.Some(*call.callers[0])
	}

	var errorStr optional.Option[string]
	var stack optional.Option[string]
	var responseBody []byte

	if callError, ok := call.callError.Get(); ok {
		errorStr = optional.Some(callError.Error())
	} else if response, ok := call.response.Get(); ok {
		responseBody = response.GetBody()
		if callError := response.GetError(); callError != nil {
			errorStr = optional.Some(callError.Message)
			stack = optional.Ptr(callError.Stack)
		}
	}

	err := s.dal.InsertCallEvent(ctx, &dal.CallEvent{
		Time:          call.startTime,
		DeploymentKey: call.deploymentKey,
		RequestKey:    optional.Some(call.requestKey),
		Duration:      time.Since(call.startTime),
		SourceVerb:    sourceVerb,
		DestVerb:      *call.destVerb,
		Request:       call.request.GetBody(),
		Response:      responseBody,
		Error:         errorStr,
		Stack:         stack,
	})
	if err != nil {
		logger.Errorf(err, "failed to record call")
	}
}
