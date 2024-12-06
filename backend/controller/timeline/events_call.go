package timeline

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/alecthomas/types/either"
	"github.com/alecthomas/types/optional"

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
