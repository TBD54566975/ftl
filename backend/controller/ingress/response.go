package ingress

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/TBD54566975/ftl/backend/schema"
)

func ResponseBodyForVerb(sch *schema.Schema, verb *schema.Verb, body []byte, headers map[string][]string) ([]byte, error) {
	responseRef, ok := verb.Response.(*schema.DataRef)
	if !ok {
		return body, nil
	}

	bodyField, err := getBodyField(responseRef, sch)
	if err != nil {
		return nil, err
	}

	switch bodyType := bodyField.Type.(type) {
	case *schema.DataRef:
		var responseMap map[string]any
		err := json.Unmarshal(body, &responseMap)
		if err != nil {
			return nil, fmt.Errorf("HTTP response body is not valid JSON: %w", err)
		}

		aliasedResponseMap, err := transformToAliasedFields(bodyType, sch, responseMap)
		if err != nil {
			return nil, err
		}
		return json.Marshal(aliasedResponseMap)

	case *schema.Bytes:
		var base64String string
		if err := json.Unmarshal(body, &base64String); err != nil {
			return nil, fmt.Errorf("HTTP response body is not valid base64: %w", err)
		}
		decodedBody, err := base64.StdEncoding.DecodeString(base64String)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 response body: %w", err)
		}
		return decodedBody, nil

	case *schema.String:
		var responseString string
		if err := json.Unmarshal(body, &responseString); err != nil {
			return nil, fmt.Errorf("HTTP response body is not valid string: %w", err)
		}
		return []byte(responseString), nil

	case *schema.Unit:
		return []byte{}, nil

	default:
		return body, nil
	}
}
