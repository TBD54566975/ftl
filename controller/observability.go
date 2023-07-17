package controller

import (
	"context"
	"github.com/TBD54566975/ftl/common/model"
	"github.com/TBD54566975/ftl/controller/internal/dal"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/schema"
	"github.com/alecthomas/errors"
	"time"
)

type callDuration struct {
	requestID     int64
	runnerKey     model.RunnerKey
	controllerKey model.ControllerKey
	startTime     time.Time
	destVerb      *schema.VerbRef
	callers       []*schema.VerbRef
	request       []byte
	response      []byte
}

type callError struct {
	requestID     int64
	runnerKey     model.RunnerKey
	controllerKey model.ControllerKey
	startTime     time.Time
	destVerb      *schema.VerbRef
	callers       []*schema.VerbRef
	request       []byte
	error         error
}

func (s *Service) recordCallDuration(ctx context.Context, callDuration *callDuration) error {
	sourceVerb := schema.VerbRef{}
	if len(callDuration.callers) > 1 {
		sourceVerb = *callDuration.callers[1]
	}
	err := s.dal.InsertCallEntry(ctx, &dal.CallEntry{
		RequestID:     callDuration.requestID,
		RunnerKey:     callDuration.runnerKey,
		ControllerKey: callDuration.controllerKey,
		Duration:      time.Since(callDuration.startTime),
		SourceVerb:    sourceVerb,
		DestVerb:      *callDuration.destVerb,
		Request:       callDuration.request,
		Response:      callDuration.response,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Service) recordCallError(ctx context.Context, callError *callError) {
	logger := log.FromContext(ctx)
	logger.Warnf("recordCallError not implemented")
}
