package timeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/alecthomas/types/either"
	"github.com/alecthomas/types/optional"

	ftlencryption "github.com/TBD54566975/ftl/backend/controller/encryption/api"
	"github.com/TBD54566975/ftl/backend/controller/timeline/internal/sql"
	"github.com/TBD54566975/ftl/backend/libdal"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
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

func (c *Call) toEvent() (Event, error) { return callToCallEvent(c), nil } //nolint:unparam

func (s *Service) insertCallEvent(ctx context.Context, querier sql.Querier, callEvent *CallEvent) error {
	var sourceModule, sourceVerb optional.Option[string]
	if sr, ok := callEvent.SourceVerb.Get(); ok {
		sourceModule, sourceVerb = optional.Some(sr.Module), optional.Some(sr.Name)
	}

	var requestKey optional.Option[string]
	if rn, ok := callEvent.RequestKey.Get(); ok {
		requestKey = optional.Some(rn.String())
	}

	var parentRequestKey optional.Option[string]
	if pr, ok := callEvent.ParentRequestKey.Get(); ok {
		parentRequestKey = optional.Some(pr.String())
	}

	callJSON := eventCallJSON{
		DurationMS: callEvent.Duration.Milliseconds(),
		Request:    callEvent.Request,
		Response:   callEvent.Response,
		Error:      callEvent.Error,
		Stack:      callEvent.Stack,
	}

	data, err := json.Marshal(callJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal call event: %w", err)
	}

	var payload ftlencryption.EncryptedTimelineColumn
	err = s.encryption.EncryptJSON(json.RawMessage(data), &payload)
	if err != nil {
		return fmt.Errorf("failed to encrypt call event: %w", err)
	}

	err = libdal.TranslatePGError(querier.InsertTimelineCallEvent(ctx, sql.InsertTimelineCallEventParams{
		DeploymentKey:    callEvent.DeploymentKey,
		RequestKey:       requestKey,
		ParentRequestKey: parentRequestKey,
		TimeStamp:        callEvent.Time,
		SourceModule:     sourceModule,
		SourceVerb:       sourceVerb,
		DestModule:       callEvent.DestVerb.Module,
		DestVerb:         callEvent.DestVerb.Name,
		Payload:          payload,
	}))
	if err != nil {
		return fmt.Errorf("failed to insert call event: %w", err)
	}
	return nil
}

func callToCallEvent(call *Call) *CallEvent {
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

	return &CallEvent{
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
	}
}

func CallEventToCallForTesting(event *CallEvent) *Call {
	var response either.Either[*ftlv1.CallResponse, error]
	if eventErr, ok := event.Error.Get(); ok {
		response = either.RightOf[*ftlv1.CallResponse](errors.New(eventErr))
	} else {
		response = either.LeftOf[error](&ftlv1.CallResponse{
			Response: &ftlv1.CallResponse_Body{
				Body: event.Response,
			},
		})
	}

	var requestKey model.RequestKey
	if key, ok := event.RequestKey.Get(); ok {
		requestKey = key
	} else {
		requestKey = model.RequestKey{}
	}

	callers := []*schema.Ref{}
	if ref, ok := event.SourceVerb.Get(); ok {
		callers = []*schema.Ref{&ref}
	}

	return &Call{
		DeploymentKey:    event.DeploymentKey,
		RequestKey:       requestKey,
		ParentRequestKey: event.ParentRequestKey,
		StartTime:        event.Time,
		DestVerb:         &event.DestVerb,
		Callers:          callers,
		Request:          &ftlv1.CallRequest{Body: event.Request},
		Response:         response,
	}
}
