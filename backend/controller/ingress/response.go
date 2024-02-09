package ingress

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"strconv"

	"github.com/TBD54566975/ftl/backend/schema"
)

// HTTPResponse mirrors builtins.HttpResponse.
type HTTPResponse struct {
	Status  int                 `json:"status,omitempty"`
	Headers map[string][]string `json:"headers,omitempty"`
	Body    json.RawMessage     `json:"body,omitempty"`
	Error   json.RawMessage     `json:"error,omitempty"`
}

// ResponseForVerb returns the HTTP response for a given verb.
func ResponseForVerb(sch *schema.Schema, verb *schema.Verb, response HTTPResponse) ([]byte, http.Header, error) {
	responseRef, ok := verb.Response.(*schema.DataRef)
	if !ok {
		return nil, nil, nil
	}

	bodyData, err := sch.ResolveDataRefMonomorphised(responseRef)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve response data type: %w", err)
	}

	haveBody := response.Body != nil && !bytes.Equal(response.Body, []byte("null"))
	haveError := response.Error != nil && !bytes.Equal(response.Error, []byte("null"))

	var fieldType schema.Type
	var body []byte

	switch {
	case haveBody == haveError:
		return nil, nil, fmt.Errorf("response must have either a body or an error")

	case haveBody:
		fieldType = bodyData.FieldByName("body").Type.(*schema.Optional).Type //nolint:forcetypeassert
		body = response.Body

	case haveError:
		fieldType = bodyData.FieldByName("error").Type.(*schema.Optional).Type //nolint:forcetypeassert
		body = response.Error
	}

	// Clone and canonicalise the headers.
	headers := http.Header(maps.Clone(response.Headers))
	for k, v := range response.Headers {
		headers[http.CanonicalHeaderKey(k)] = v
	}
	// If the Content-Type header is not set, set it to the default value for the response or error type.
	if _, ok := headers["Content-Type"]; !ok {
		if contentType := getDefaultContentType(fieldType); contentType != "" {
			headers.Set("Content-Type", getDefaultContentType(fieldType))
		}
	}

	switch bodyType := fieldType.(type) {
	case *schema.DataRef:
		var responseMap map[string]any
		err := json.Unmarshal(body, &responseMap)
		if err != nil {
			return nil, nil, fmt.Errorf("HTTP response body is not valid JSON: %w", err)
		}

		aliasedResponseMap, err := transformToAliasedFields(bodyType, sch, responseMap)
		if err != nil {
			return nil, nil, err
		}
		outBody, err := json.Marshal(aliasedResponseMap)
		return outBody, headers, err

	case *schema.Bytes:
		var base64String string
		if err := json.Unmarshal(body, &base64String); err != nil {
			return nil, nil, fmt.Errorf("HTTP response body is not valid base64: %w", err)
		}
		decodedBody, err := base64.StdEncoding.DecodeString(base64String)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode base64 response body: %w", err)
		}
		return decodedBody, headers, nil

	case *schema.String:
		var responseString string
		if err := json.Unmarshal(body, &responseString); err != nil {
			return nil, nil, fmt.Errorf("HTTP response body is not a valid string: %w", err)
		}
		return []byte(responseString), headers, nil

	case *schema.Int:
		var responseInt int
		if err := json.Unmarshal(body, &responseInt); err != nil {
			return nil, nil, fmt.Errorf("HTTP response body is not a valid int: %w", err)
		}
		return []byte(strconv.Itoa(responseInt)), headers, nil

	case *schema.Float:
		var responseFloat float64
		if err := json.Unmarshal(body, &responseFloat); err != nil {
			return nil, nil, fmt.Errorf("HTTP response body is not a valid float: %w", err)
		}
		return []byte(strconv.FormatFloat(responseFloat, 'f', -1, 64)), headers, nil

	case *schema.Bool:
		var responseBool bool
		if err := json.Unmarshal(body, &responseBool); err != nil {
			return nil, nil, fmt.Errorf("HTTP response body is not a valid bool: %w", err)
		}
		return []byte(strconv.FormatBool(responseBool)), headers, nil

	case *schema.Unit:
		return []byte{}, headers, nil

	default:
		return body, headers, nil
	}
}

func getDefaultContentType(typ schema.Type) string {
	switch typ.(type) {
	case *schema.Bytes:
		return "application/octet-stream"
	case *schema.String, *schema.Int, *schema.Float, *schema.Bool:
		return "text/plain; charset=utf-8"
	case *schema.DataRef, *schema.Map, *schema.Array:
		return "application/json; charset=utf-8"
	default:
		return ""
	}
}
