package timeline

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
)

type IngressEvent struct {
	ID            int64
	DeploymentKey model.DeploymentKey
	RequestKey    optional.Option[model.RequestKey]
	Verb          schema.Ref
	Method        string
	Path          string

	StatusCode     int
	Time           time.Time
	Duration       time.Duration
	Request        json.RawMessage
	RequestHeader  json.RawMessage
	Response       json.RawMessage
	ResponseHeader json.RawMessage
	Error          optional.Option[string]
}

func (e *IngressEvent) GetID() int64 { return e.ID }
func (e *IngressEvent) event()       {}

type eventIngressJSON struct {
	DurationMS     int64                   `json:"duration_ms"`
	Method         string                  `json:"method"`
	Path           string                  `json:"path"`
	StatusCode     int                     `json:"status_code"`
	Request        json.RawMessage         `json:"request"`
	RequestHeader  json.RawMessage         `json:"req_header"`
	Response       json.RawMessage         `json:"response"`
	ResponseHeader json.RawMessage         `json:"resp_header"`
	Error          optional.Option[string] `json:"error,omitempty"`
}

type Ingress struct {
	DeploymentKey   model.DeploymentKey
	RequestKey      model.RequestKey
	StartTime       time.Time
	Verb            *schema.Ref
	RequestMethod   string
	RequestPath     string
	RequestHeaders  http.Header
	ResponseStatus  int
	ResponseHeaders http.Header
	RequestBody     []byte
	ResponseBody    []byte
	Error           optional.Option[string]
}

func (ingress *Ingress) toEvent() (Event, error) {
	requestBody := ingress.RequestBody
	if len(requestBody) == 0 {
		requestBody = []byte("{}")
	}

	responseBody := ingress.ResponseBody
	if len(responseBody) == 0 {
		responseBody = []byte("{}")
	}

	reqHeaderBytes, err := json.Marshal(ingress.RequestHeaders)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request header: %w", err)
	}

	respHeaderBytes, err := json.Marshal(ingress.ResponseHeaders)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response header: %w", err)
	}
	return &IngressEvent{
		DeploymentKey:  ingress.DeploymentKey,
		RequestKey:     optional.Some(ingress.RequestKey),
		Verb:           *ingress.Verb,
		Method:         ingress.RequestMethod,
		Path:           ingress.RequestPath,
		StatusCode:     ingress.ResponseStatus,
		Time:           ingress.StartTime,
		Duration:       time.Since(ingress.StartTime),
		Request:        requestBody,
		RequestHeader:  reqHeaderBytes,
		Response:       responseBody,
		ResponseHeader: respHeaderBytes,
		Error:          ingress.Error,
	}, nil
}
