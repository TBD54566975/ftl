package ingress

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/controller/observability"
	dalerrs "github.com/TBD54566975/ftl/backend/dal"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
)

// Handle HTTP ingress routes.
func Handle(
	startTime time.Time,
	sch *schema.Schema,
	requestKey model.RequestKey,
	routes []dal.IngressRoute,
	w http.ResponseWriter,
	r *http.Request,
	call func(context.Context, *connect.Request[ftlv1.CallRequest], optional.Option[model.RequestKey], optional.Option[model.RequestKey], string) (*connect.Response[ftlv1.CallResponse], error),
) {
	logger := log.FromContext(r.Context())
	logger.Debugf("%s %s", r.Method, r.URL.Path)
	route, err := GetIngressRoute(routes, r.Method, r.URL.Path)
	if err != nil {
		if errors.Is(err, dalerrs.ErrNotFound) {
			http.NotFound(w, r)
			observability.Ingress.Request(r.Context(), r.Method, r.URL.Path, optional.None[*schemapb.Ref](), startTime, optional.Some("route not found"))
			return
		}
		logger.Errorf(err, "failed to resolve route for %s %s", r.Method, r.URL.Path)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		observability.Ingress.Request(r.Context(), r.Method, r.URL.Path, optional.None[*schemapb.Ref](), startTime, optional.Some("failed to resolve route"))
		return
	}

	verbRef := &schemapb.Ref{Module: route.Module, Name: route.Verb}

	body, err := BuildRequestBody(route, r, sch)
	if err != nil {
		// Only log at debug, as this is a client side error
		logger.Debugf("bad request: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		observability.Ingress.Request(r.Context(), r.Method, r.URL.Path, optional.Some(verbRef), startTime, optional.Some("bad request"))
		return
	}

	creq := connect.NewRequest(&ftlv1.CallRequest{
		Metadata: &ftlv1.Metadata{},
		Verb:     verbRef,
		Body:     body,
	})

	resp, err := call(r.Context(), creq, optional.Some(requestKey), optional.None[model.RequestKey](), r.RemoteAddr)
	if err != nil {
		logger.Errorf(err, "failed to call verb %s", route.Verb)
		if connectErr := new(connect.Error); errors.As(err, &connectErr) {
			httpCode := connectCodeToHTTP(connectErr.Code())
			http.Error(w, http.StatusText(httpCode), httpCode)
			observability.Ingress.Request(r.Context(), r.Method, r.URL.Path, optional.Some(verbRef), startTime, optional.Some("failed to call verb: connect error"))
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			observability.Ingress.Request(r.Context(), r.Method, r.URL.Path, optional.Some(verbRef), startTime, optional.Some("failed to call verb: internal server error"))
		}
		return
	}
	switch msg := resp.Msg.Response.(type) {
	case *ftlv1.CallResponse_Body:
		verb := &schema.Verb{}
		err = sch.ResolveToType(&schema.Ref{Name: route.Verb, Module: route.Module}, verb)
		if err != nil {
			logger.Errorf(err, "could not resolve schema type for verb %s", route.Verb)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			observability.Ingress.Request(r.Context(), r.Method, r.URL.Path, optional.Some(verbRef), startTime, optional.Some("could not resolve schema type for verb"))
			return
		}
		var responseBody []byte

		if metadata, ok := verb.GetMetadataIngress().Get(); ok && metadata.Type == "http" {
			var response HTTPResponse
			if err := json.Unmarshal(msg.Body, &response); err != nil {
				logger.Errorf(err, "could not unmarhal response for verb %s", verb)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				observability.Ingress.Request(r.Context(), r.Method, r.URL.Path, optional.Some(verbRef), startTime, optional.Some("could not unmarhal response for verb"))
				return
			}

			var responseHeaders http.Header
			responseBody, responseHeaders, err = ResponseForVerb(sch, verb, response)
			if err != nil {
				logger.Errorf(err, "could not create response for verb %s", verb)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				observability.Ingress.Request(r.Context(), r.Method, r.URL.Path, optional.Some(verbRef), startTime, optional.Some("could not create response for verb"))
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
		if err == nil {
			observability.Ingress.Request(r.Context(), r.Method, r.URL.Path, optional.Some(verbRef), startTime, optional.None[string]())
		} else {
			logger.Errorf(err, "Could not write response body")
			observability.Ingress.Request(r.Context(), r.Method, r.URL.Path, optional.Some(verbRef), startTime, optional.Some("could not write response body"))
		}

	case *ftlv1.CallResponse_Error_:
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		observability.Ingress.Request(r.Context(), r.Method, r.URL.Path, optional.Some(verbRef), startTime, optional.Some("call response: internal server error"))
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
