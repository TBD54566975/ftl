package headers

import (
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
)

// Headers used by the internal RPC system.
const (
	DirectRoutingHeader = "Ftl-Direct"
	// VerbHeader is the header used to pass the module.verb of the current request.
	//
	// One header will be present for each hop in the request path.
	VerbHeader = "Ftl-Verb"
	// RequestIDHeader is the header used to pass the inbound request ID.
	RequestIDHeader = "Ftl-Request-Id"
	// ParentRequestIDHeader is the header used to pass the parent request ID,
	// i.e. the publisher that initiated this call.
	ParentRequestIDHeader = "Ftl-Parent-Request-Id"

	transferEncoding = "Transfer-Encoding"
	headerHost       = "Host"
)

var protocolHeaders = map[string]bool{
	headerHost:       true,
	transferEncoding: true,
}

// CopyRequestForForwarding creates a new request with the same message as the original, but with only the FTL specific headers
func CopyRequestForForwarding[T any](req *connect.Request[T]) *connect.Request[T] {
	ret := connect.NewRequest(req.Msg)
	for key, val := range req.Header() {
		if !protocolHeaders[key] {
			ret.Header()[key] = val
		}
	}
	return ret
}

func IsDirectRouted(header http.Header) bool {
	return header.Get(DirectRoutingHeader) != ""
}

func SetDirectRouted(header http.Header) {
	header.Set(DirectRoutingHeader, "1")
}

func SetRequestKey(header http.Header, key model.RequestKey) {
	header.Set(RequestIDHeader, key.String())
}

func SetParentRequestKey(header http.Header, key model.RequestKey) {
	header.Set(ParentRequestIDHeader, key.String())
}

// GetRequestKey from an incoming request.
//
// Will return ("", false, nil) if no request key is present.
func GetRequestKey(header http.Header) (model.RequestKey, bool, error) {
	keyStr := header.Get(RequestIDHeader)
	return getRequestKeyFromKeyStr(keyStr)
}

func GetParentRequestKey(header http.Header) (model.RequestKey, bool, error) {
	keyStr := header.Get(ParentRequestIDHeader)
	return getRequestKeyFromKeyStr(keyStr)
}

func getRequestKeyFromKeyStr(keyStr string) (model.RequestKey, bool, error) {
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
