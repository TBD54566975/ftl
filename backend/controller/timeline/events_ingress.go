package timeline

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alecthomas/types/optional"

	ftlencryption "github.com/TBD54566975/ftl/backend/controller/encryption/api"
	"github.com/TBD54566975/ftl/backend/controller/timeline/internal/sql"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/internal/log"
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

func (*Ingress) inEvent() {}

func (s *Service) insertHTTPIngress(ctx context.Context, querier sql.Querier, ingress *Ingress) error {
	requestBody := ingress.RequestBody
	if len(requestBody) == 0 {
		requestBody = []byte("{}")
	}

	responseBody := ingress.ResponseBody
	if len(responseBody) == 0 {
		responseBody = []byte("{}")
	}

	if len(responseBody) == 0 {
		responseBody = []byte("{}")
	}

	reqHeaderBytes, err := json.Marshal(ingress.RequestHeaders)
	if err != nil {
		return fmt.Errorf("failed to marshal request header: %w", err)
	}

	respHeaderBytes, err := json.Marshal(ingress.ResponseHeaders)
	if err != nil {
		return fmt.Errorf("failed to marshal response header: %w", err)
	}

	ingressJSON := eventIngressJSON{
		DurationMS:     time.Since(ingress.StartTime).Milliseconds(),
		Method:         ingress.RequestMethod,
		Path:           ingress.RequestPath,
		StatusCode:     ingress.ResponseStatus,
		Request:        json.RawMessage(requestBody),
		RequestHeader:  json.RawMessage(reqHeaderBytes),
		Response:       json.RawMessage(responseBody),
		ResponseHeader: json.RawMessage(respHeaderBytes),
		Error:          ingress.Error,
	}

	data, err := json.Marshal(ingressJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal ingress JSON: %w", err)
	}

	var payload ftlencryption.EncryptedTimelineColumn
	err = s.encryption.EncryptJSON(json.RawMessage(data), &payload)
	if err != nil {
		return fmt.Errorf("failed to encrypt ingress payload: %w", err)
	}

	log.FromContext(ctx).Warnf("Inserting ingress event for %s %s", ingress.RequestKey, ingress.RequestPath)

	err = libdal.TranslatePGError(querier.InsertTimelineIngressEvent(ctx, sql.InsertTimelineIngressEventParams{
		DeploymentKey: ingress.DeploymentKey,
		RequestKey:    optional.Some(ingress.RequestKey.String()),
		TimeStamp:     ingress.StartTime,
		Module:        ingress.Verb.Module,
		Verb:          ingress.Verb.Name,
		IngressType:   "http",
		Payload:       payload,
	}))
	if err != nil {
		return fmt.Errorf("failed to insert ingress event: %w", err)
	}
	return nil
}
