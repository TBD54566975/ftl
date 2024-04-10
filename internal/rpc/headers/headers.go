package headers

import (
	"fmt"
	"net/http"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/model"
)

// Headers used by the internal RPC system.
const (
	DirectRoutingHeader = "FTL-Direct"
	// VerbHeader is the header used to pass the module.verb of the current request.
	//
	// One header will be present for each hop in the request path.
	VerbHeader = "FTL-Verb"
	// RequestIDHeader is the header used to pass the inbound request ID.
	RequestIDHeader = "FTL-Request-ID"
)

func IsDirectRouted(header http.Header) bool {
	return header.Get(DirectRoutingHeader) != ""
}

func SetDirectRouted(header http.Header) {
	header.Set(DirectRoutingHeader, "1")
}

func SetRequestName(header http.Header, key model.RequestKey) {
	header.Set(RequestIDHeader, key.String())
}

// GetRequestName from an incoming request.
//
// Will return ("", false, nil) if no request key is present.
func GetRequestName(header http.Header) (model.RequestKey, bool, error) {
	keyStr := header.Get(RequestIDHeader)
	if keyStr == "" {
		return model.RequestKey{}, false, nil
	}

	key, err := model.ParseRequestKey(keyStr)
	if err != nil {
		return model.RequestKey{}, false, err
	}
	return key, true, nil
}

// GetCallers history from an incoming request.
func GetCallers(header http.Header) ([]*schema.Ref, error) {
	headers := header.Values(VerbHeader)
	if len(headers) == 0 {
		return nil, nil
	}
	refs := make([]*schema.Ref, len(headers))
	for i, header := range headers {
		ref, err := schema.ParseRef(header)
		if err != nil {
			return nil, fmt.Errorf("invalid %s header %q: %w", VerbHeader, header, err)
		}
		refs[i] = ref
	}
	return refs, nil
}

// GetCaller returns the module.verb of the caller, if any.
//
// Will return an error if the header is malformed.
func GetCaller(header http.Header) (optional.Option[*schema.Ref], error) {
	headers := header.Values(VerbHeader)
	if len(headers) == 0 {
		return optional.None[*schema.Ref](), nil
	}
	ref, err := schema.ParseRef(headers[len(headers)-1])
	if err != nil {
		return optional.None[*schema.Ref](), err
	}
	return optional.Some(ref), nil
}

// AddCaller to an outgoing request.
func AddCaller(header http.Header, ref *schema.Ref) {
	refStr := ref.String()
	if values := header.Values(VerbHeader); len(values) > 0 {
		if values[len(values)-1] == refStr {
			return
		}
	}
	header.Add(VerbHeader, refStr)
}

func SetCallers(header http.Header, refs []*schema.Ref) {
	header.Del(VerbHeader)
	for _, ref := range refs {
		AddCaller(header, ref)
	}
}
