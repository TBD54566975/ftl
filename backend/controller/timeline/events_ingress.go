package timeline

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/alecthomas/types/optional"

	ftlencryption "github.com/TBD54566975/ftl/backend/controller/encryption/api"
	"github.com/TBD54566975/ftl/backend/controller/timeline/internal/sql"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
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
	DeploymentKey model.DeploymentKey
	RequestKey    model.RequestKey
	StartTime     time.Time
	Verb          *schema.Ref
	Request       *http.Request
	Response      *http.Response
	Error         optional.Option[string]
}

func (s *Service) InsertHTTPIngress(ctx context.Context, ingress *Ingress) {
	logger := log.FromContext(ctx)

	requestBody, err := io.ReadAll(ingress.Request.Body)
	if err != nil {
		logger.Errorf(err, "failed to read request body")
		return
	}
	if len(requestBody) == 0 {
		requestBody = []byte("{}")
	}

	var responseBody []byte
	if ingress.Response.Body != nil {
		responseBody, err = io.ReadAll(ingress.Response.Body)
		if err != nil {
			logger.Errorf(err, "failed to read response body")
			return
		}
	}

	if len(responseBody) == 0 {
		responseBody = []byte("{}")
	}

	reqHeaderBytes, err := json.Marshal(ingress.Request.Header)
	if err != nil {
		logger.Errorf(err, "failed to marshal request header")
		return
	}
	if len(reqHeaderBytes) == 0 {
		reqHeaderBytes = []byte("{}")
	}

	respHeaderBytes, err := json.Marshal(ingress.Response.Header)
	if err != nil {
		logger.Errorf(err, "failed to marshal response header")
		return
	}
	if len(respHeaderBytes) == 0 {
		respHeaderBytes = []byte("{}")
	}

	ingressJSON := eventIngressJSON{
		DurationMS:     time.Since(ingress.StartTime).Milliseconds(),
		Method:         ingress.Request.Method,
		Path:           ingress.Request.URL.Path,
		StatusCode:     ingress.Response.StatusCode,
		Request:        json.RawMessage(requestBody),
		RequestHeader:  json.RawMessage(reqHeaderBytes),
		Response:       json.RawMessage(responseBody),
		ResponseHeader: json.RawMessage(respHeaderBytes),
		Error:          ingress.Error,
	}

	data, err := json.Marshal(ingressJSON)
	if err != nil {
		logger.Errorf(err, "failed to marshal ingress JSON")
		return
	}

	var payload ftlencryption.EncryptedTimelineColumn
	err = s.encryption.EncryptJSON(json.RawMessage(data), &payload)
	if err != nil {
		logger.Errorf(err, "failed to encrypt ingress payload")
		return
	}

	err = libdal.TranslatePGError(s.db.InsertTimelineIngressEvent(ctx, sql.InsertTimelineIngressEventParams{
		DeploymentKey: ingress.DeploymentKey,
		RequestKey:    optional.Some(ingress.RequestKey.String()),
		TimeStamp:     ingress.StartTime,
		Module:        ingress.Verb.Module,
		Verb:          ingress.Verb.Name,
		IngressType:   "http",
		Payload:       payload,
	}))
	if err != nil {
		logger.Errorf(err, "failed to insert ingress event")
	}
}
