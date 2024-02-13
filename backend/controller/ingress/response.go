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

	outBody, err := bodyForType(fieldType, sch, body)
	return outBody, headers, err
}

func bodyForType(typ schema.Type, sch *schema.Schema, data []byte) ([]byte, error) {
	switch t := typ.(type) {
	case *schema.DataRef, *schema.Array, *schema.Map:
		var response any
		err := json.Unmarshal(data, &response)
		if err != nil {
			return nil, fmt.Errorf("HTTP response body is not valid JSON: %w", err)
		}

		err = transformAliasedFields(sch, t, response, func(obj map[string]any, field *schema.Field) string {
			if field.Alias != "" && field.Name != field.Alias {
				obj[field.Alias] = obj[field.Name]
				delete(obj, field.Name)
				return field.Alias
			}
			return field.Name
		})
		if err != nil {
			return nil, err
		}
		outBody, err := json.Marshal(response)
		return outBody, err

	case *schema.Bytes:
		var base64String string
		if err := json.Unmarshal(data, &base64String); err != nil {
			return nil, fmt.Errorf("HTTP response body is not valid base64: %w", err)
		}
		decodedBody, err := base64.StdEncoding.DecodeString(base64String)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 response body: %w", err)
		}
		return decodedBody, nil

	case *schema.String:
		var responseString string
		if err := json.Unmarshal(data, &responseString); err != nil {
			return nil, fmt.Errorf("HTTP response body is not a valid string: %w", err)
		}
		return []byte(responseString), nil

	case *schema.Int:
		var responseInt int
		if err := json.Unmarshal(data, &responseInt); err != nil {
			return nil, fmt.Errorf("HTTP response body is not a valid int: %w", err)
		}
		return []byte(strconv.Itoa(responseInt)), nil

	case *schema.Float:
		var responseFloat float64
		if err := json.Unmarshal(data, &responseFloat); err != nil {
			return nil, fmt.Errorf("HTTP response body is not a valid float: %w", err)
		}
		return []byte(strconv.FormatFloat(responseFloat, 'f', -1, 64)), nil

	case *schema.Bool:
		var responseBool bool
		if err := json.Unmarshal(data, &responseBool); err != nil {
			return nil, fmt.Errorf("HTTP response body is not a valid bool: %w", err)
		}
		return []byte(strconv.FormatBool(responseBool)), nil

	case *schema.Unit:
		return []byte{}, nil

	default:
		return data, nil
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
