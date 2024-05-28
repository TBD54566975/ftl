package ingress

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/controller/dalerrors"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
)

// Handle HTTP ingress routes.
func Handle(
	sch *schema.Schema,
	requestKey model.RequestKey,
	routes []dal.IngressRoute,
	w http.ResponseWriter,
	r *http.Request,
	call func(context.Context, *connect.Request[ftlv1.CallRequest], optional.Option[model.RequestKey], string) (*connect.Response[ftlv1.CallResponse], error),
) {
	logger := log.FromContext(r.Context())
	logger.Debugf("%s %s", r.Method, r.URL.Path)
	route, err := GetIngressRoute(routes, r.Method, r.URL.Path)
	if err != nil {
		if errors.Is(err, dalerrors.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := BuildRequestBody(route, r, sch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	creq := connect.NewRequest(&ftlv1.CallRequest{
		Metadata: &ftlv1.Metadata{},
		Verb:     &schemapb.Ref{Module: route.Module, Name: route.Verb},
		Body:     body,
	})

	resp, err := call(r.Context(), creq, optional.Some(requestKey), r.RemoteAddr)
	if err != nil {
		if connectErr := new(connect.Error); errors.As(err, &connectErr) {
			http.Error(w, err.Error(), connectCodeToHTTP(connectErr.Code()))
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	switch msg := resp.Msg.Response.(type) {
	case *ftlv1.CallResponse_Body:
		verb := &schema.Verb{}
		err = sch.ResolveToType(&schema.Ref{Name: route.Verb, Module: route.Module}, verb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var responseBody []byte

		if metadata, ok := verb.GetMetadataIngress().Get(); ok && metadata.Type == "http" {
			var response HTTPResponse
			if err := json.Unmarshal(msg.Body, &response); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var responseHeaders http.Header
			responseBody, responseHeaders, err = ResponseForVerb(sch, verb, response)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			for k, v := range responseHeaders {
				w.Header()[k] = v
			}

			if response.Status != 0 {
				w.WriteHeader(response.Status)
			}
		} else {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			responseBody = msg.Body
		}
		_, err = w.Write(responseBody)
		if err != nil {
			logger.Errorf(err, "Could not write response body")
		}

	case *ftlv1.CallResponse_Error_:
		http.Error(w, msg.Error.Message, http.StatusInternalServerError)
	}
}

// Copied from the Apache-licensed connect-go source.
func connectCodeToHTTP(code connect.Code) int {
	switch code {
	case connect.CodeCanceled:
		return 408
	case connect.CodeUnknown:
		return 500
	case connect.CodeInvalidArgument:
		return 400
	case connect.CodeDeadlineExceeded:
		return 408
	case connect.CodeNotFound:
		return 404
	case connect.CodeAlreadyExists:
		return 409
	case connect.CodePermissionDenied:
		return 403
	case connect.CodeResourceExhausted:
		return 429
	case connect.CodeFailedPrecondition:
		return 412
	case connect.CodeAborted:
		return 409
	case connect.CodeOutOfRange:
		return 400
	case connect.CodeUnimplemented:
		return 404
	case connect.CodeInternal:
		return 500
	case connect.CodeUnavailable:
		return 503
	case connect.CodeDataLoss:
		return 500
	case connect.CodeUnauthenticated:
		return 401
	default:
		return 500 // same as CodeUnknown
	}
}
