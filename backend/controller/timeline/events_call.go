package timeline

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	ftlencryption "github.com/TBD54566975/ftl/backend/controller/encryption/api"
	"github.com/TBD54566975/ftl/backend/controller/timeline/internal/sql"
	"github.com/TBD54566975/ftl/backend/libdal"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/alecthomas/types/either"
	"github.com/alecthomas/types/optional"
)

type CallEvent struct {
	ID               int64
	DeploymentKey    model.DeploymentKey
	RequestKey       optional.Option[model.RequestKey]
	ParentRequestKey optional.Option[model.RequestKey]
	Time             time.Time
	SourceVerb       optional.Option[schema.Ref]
	DestVerb         schema.Ref
	Duration         time.Duration
	Request          json.RawMessage
	Response         json.RawMessage
	Error            optional.Option[string]
	Stack            optional.Option[string]
}

func (e *CallEvent) GetID() int64 { return e.ID }
func (e *CallEvent) event()       {}

// The internal JSON payload of a call event.
type eventCallJSON struct {
	DurationMS int64                   `json:"duration_ms"`
	Request    json.RawMessage         `json:"request"`
	Response   json.RawMessage         `json:"response"`
	Error      optional.Option[string] `json:"error,omitempty"`
	Stack      optional.Option[string] `json:"stack,omitempty"`
}

type Call struct {
	DeploymentKey    model.DeploymentKey
	RequestKey       model.RequestKey
	ParentRequestKey optional.Option[model.RequestKey]
	StartTime        time.Time
	DestVerb         *schema.Ref
	Callers          []*schema.Ref
	Request          *ftlv1.CallRequest
	Response         either.Either[*ftlv1.CallResponse, error]
}

func (s *Service) RecordCall(ctx context.Context, call *Call) {
	logger := log.FromContext(ctx)
	var sourceVerb optional.Option[schema.Ref]
	if len(call.Callers) > 0 {
		sourceVerb = optional.Some(*call.Callers[0])
	}

	var errorStr optional.Option[string]
	var stack optional.Option[string]
	var responseBody []byte

	switch response := call.Response.(type) {
	case either.Left[*ftlv1.CallResponse, error]:
		resp := response.Get()
		responseBody = resp.GetBody()
		if callError := resp.GetError(); callError != nil {
			errorStr = optional.Some(callError.Message)
			stack = optional.Ptr(callError.Stack)
		}
	case either.Right[*ftlv1.CallResponse, error]:
		callError := response.Get()
		errorStr = optional.Some(callError.Error())
	}

	err := s.insertCallEvent(ctx, &CallEvent{
		Time:             call.StartTime,
		DeploymentKey:    call.DeploymentKey,
		RequestKey:       optional.Some(call.RequestKey),
		ParentRequestKey: call.ParentRequestKey,
		Duration:         time.Since(call.StartTime),
		SourceVerb:       sourceVerb,
		DestVerb:         *call.DestVerb,
		Request:          call.Request.GetBody(),
		Response:         responseBody,
		Error:            errorStr,
		Stack:            stack,
	})
	if err != nil {
		logger.Errorf(err, "failed to record call")
	}
}

func (s *Service) insertCallEvent(ctx context.Context, call *CallEvent) error {
	var sourceModule, sourceVerb optional.Option[string]
	if sr, ok := call.SourceVerb.Get(); ok {
		sourceModule, sourceVerb = optional.Some(sr.Module), optional.Some(sr.Name)
	}
	var requestKey optional.Option[string]
	if rn, ok := call.RequestKey.Get(); ok {
		requestKey = optional.Some(rn.String())
	}
	var parentRequestKey optional.Option[string]
	if pr, ok := call.ParentRequestKey.Get(); ok {
		parentRequestKey = optional.Some(pr.String())
	}
	var payload ftlencryption.EncryptedTimelineColumn
	err := s.encryption.EncryptJSON(map[string]any{
		"duration_ms": call.Duration.Milliseconds(),
		"request":     call.Request,
		"response":    call.Response,
		"error":       call.Error,
		"stack":       call.Stack,
	}, &payload)
	if err != nil {
		return fmt.Errorf("failed to encrypt call payload: %w", err)
	}
	return libdal.TranslatePGError(s.db.InsertTimelineCallEvent(ctx, sql.InsertTimelineCallEventParams{
		DeploymentKey:    call.DeploymentKey,
		RequestKey:       requestKey,
		ParentRequestKey: parentRequestKey,
		TimeStamp:        call.Time,
		SourceModule:     sourceModule,
		SourceVerb:       sourceVerb,
		DestModule:       call.DestVerb.Module,
		DestVerb:         call.DestVerb.Name,
		Payload:          payload,
	}))
}
