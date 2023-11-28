package controller

import (
	"context"
	"time"

	"github.com/alecthomas/types"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/model"
	"github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/schema"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
)

type Call struct {
	deploymentName model.DeploymentName
	requestName    model.RequestName
	startTime      time.Time
	destVerb       *schema.VerbRef
	callers        []*schema.VerbRef
	request        *ftlv1.CallRequest
	response       types.Option[*ftlv1.CallResponse]
	callError      types.Option[error]
}

func (s *Service) recordCall(ctx context.Context, call *Call) {
	logger := log.FromContext(ctx)
	var sourceVerb types.Option[schema.VerbRef]
	if len(call.callers) > 0 {
		sourceVerb = types.Some(*call.callers[0])
	}

	var errorStr types.Option[string]
	var stack types.Option[string]
	var responseBody []byte

	if callError, ok := call.callError.Get(); ok {
		errorStr = types.Some(callError.Error())
	} else if response, ok := call.response.Get(); ok {
		responseBody = response.GetBody()
		if callError := response.GetError(); callError != nil {
			errorStr = types.Some(callError.Message)
			stack = types.Ptr(callError.Stack)
		}
	}

	err := s.dal.InsertCallEvent(ctx, &dal.CallEvent{
		Time:           call.startTime,
		DeploymentName: call.deploymentName,
		RequestName:    types.Some(call.requestName),
		Duration:       time.Since(call.startTime),
		SourceVerb:     sourceVerb,
		DestVerb:       *call.destVerb,
		Request:        call.request.GetBody(),
		Response:       responseBody,
		Error:          errorStr,
		Stack:          stack,
	})
	if err != nil {
		logger.Errorf(err, "failed to record call")
	}
}
