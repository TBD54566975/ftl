package ingress

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/schema"
)

// BuildRequestBody extracts the HttpRequest body from an HTTP request.
func BuildRequestBody(route *dal.IngressRoute, r *http.Request, sch *schema.Schema) ([]byte, error) {
	verb := sch.ResolveVerbRef(&schema.VerbRef{Name: route.Verb, Module: route.Module})
	if verb == nil {
		return nil, fmt.Errorf("unknown verb %s", route.Verb)
	}

	request, ok := verb.Request.(*schema.DataRef)
	if !ok {
		return nil, fmt.Errorf("verb %s input must be a data structure", verb.Name)
	}

	var body []byte

	var requestMap map[string]any

	if metadata, ok := verb.GetMetadataIngress().Get(); ok && metadata.Type == "http" {
		pathParameters := map[string]string{}
		matchSegments(route.Path, r.URL.Path, func(segment, value string) {
			pathParameters[segment] = value
		})

		httpRequestBody, err := extractHTTPRequestBody(route, r, request, sch)
		if err != nil {
			return nil, err
		}

		requestMap = map[string]any{}
		requestMap["method"] = r.Method
		requestMap["path"] = r.URL.Path
		requestMap["pathParameters"] = pathParameters
		requestMap["query"] = r.URL.Query()
		requestMap["headers"] = r.Header
		requestMap["body"] = httpRequestBody
	} else {
		var err error
		requestMap, err = buildRequestMap(route, r, request, sch)
		if err != nil {
			return nil, err
		}
	}

	requestMap, err := transformFromAliasedFields(request, sch, requestMap)
	if err != nil {
		return nil, err
	}

	err = validateRequestMap(request, []string{request.String()}, requestMap, sch)
	if err != nil {
		return nil, err
	}

	body, err = json.Marshal(requestMap)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func extractHTTPRequestBody(route *dal.IngressRoute, r *http.Request, dataRef *schema.DataRef, sch *schema.Schema) (any, error) {
	bodyField, err := getBodyField(dataRef, sch)
	if err != nil {
		return nil, err
	}

	switch bodyType := bodyField.Type.(type) {
	case *schema.DataRef:
		bodyMap, err := buildRequestMap(route, r, bodyType, sch)
		if err != nil {
			return nil, err
		}
		return bodyMap, nil

	case *schema.Bytes:
		bodyData, err := readRequestBody(r)
		if err != nil {
			return nil, err
		}
		return bodyData, nil

	case *schema.String:
		bodyData, err := readRequestBody(r)
		if err != nil {
			return nil, err
		}
		return string(bodyData), nil

	case *schema.Int:
		bodyData, err := readRequestBody(r)
		if err != nil {
			return nil, err
		}

		intVal, err := strconv.ParseInt(string(bodyData), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse integer from request body: %w", err)
		}
		return intVal, nil

	case *schema.Float:
		bodyData, err := readRequestBody(r)
		if err != nil {
			return nil, err
		}

		floatVal, err := strconv.ParseFloat(string(bodyData), 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse float from request body: %w", err)
		}
		return floatVal, nil

	case *schema.Bool:
		bodyData, err := readRequestBody(r)
		if err != nil {
			return nil, err
		}
		boolVal, err := strconv.ParseBool(string(bodyData))
		if err != nil {
			return nil, fmt.Errorf("failed to parse boolean from request body: %w", err)
		}
		return boolVal, nil

	case *schema.Unit:
		return map[string]any{}, nil

	default:
		return nil, fmt.Errorf("unsupported HttpRequest.Body type %T", bodyField.Type)
	}
}

func readRequestBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %w", err)
	}
	return bodyData, nil
}

func buildRequestMap(route *dal.IngressRoute, r *http.Request, dataRef *schema.DataRef, sch *schema.Schema) (map[string]any, error) {
	requestMap := map[string]any{}
	matchSegments(route.Path, r.URL.Path, func(segment, value string) {
		requestMap[segment] = value
	})

	switch r.Method {
	case http.MethodPost, http.MethodPut:
		var bodyMap map[string]any
		err := json.NewDecoder(r.Body).Decode(&bodyMap)
		if err != nil {
			return nil, fmt.Errorf("HTTP request body is not valid JSON: %w", err)
		}

		// Merge bodyMap into params
		for k, v := range bodyMap {
			requestMap[k] = v
		}
	default:
		data, err := sch.ResolveDataRefMonomorphised(dataRef)
		if err != nil {
			return nil, err
		}

		queryMap, err := parseQueryParams(r.URL.Query(), data)
		if err != nil {
			return nil, fmt.Errorf("HTTP query params are not valid: %w", err)
		}

		for key, value := range queryMap {
			requestMap[key] = value
		}
	}

	return requestMap, nil
}

func validateRequestMap(dataRef *schema.DataRef, path path, request map[string]any, sch *schema.Schema) error {
	data, err := sch.ResolveDataRefMonomorphised(dataRef)
	if err != nil {
		return err
	}

	var errs []error
	for _, field := range data.Fields {
		fieldPath := append(path, "."+field.Name) //nolint:gocritic

		_, isOptional := field.Type.(*schema.Optional)
		value, haveValue := request[field.Name]
		if !isOptional && !haveValue {
			errs = append(errs, fmt.Errorf("%s is required", fieldPath))
			continue
		}

		if haveValue {
			err := validateValue(field.Type, fieldPath, value, sch)
			if err != nil {
				errs = append(errs, err)
			}
		}

	}

	return errors.Join(errs...)
}

func parseQueryParams(values url.Values, data *schema.Data) (map[string]any, error) {
	if jsonStr, ok := values["@json"]; ok {
		if len(values) > 1 {
			return nil, fmt.Errorf("only '@json' parameter is allowed, but other parameters were found")
		}
		if len(jsonStr) > 1 {
			return nil, fmt.Errorf("'@json' parameter must be provided exactly once")
		}

		return decodeQueryJSON(jsonStr[0])
	}

	queryMap := make(map[string]any)
	for key, value := range values {
		if hasInvalidQueryChars(key) {
			return nil, fmt.Errorf("complex key %q is not supported, use '@json=' instead", key)
		}

		var field *schema.Field
		for _, f := range data.Fields {
			if (f.Alias != "" && f.Alias == key) || f.Name == key {
				field = f
			}
			for _, typeParam := range data.TypeParameters {
				if typeParam.Name == key {
					field = &schema.Field{
						Name: key,
						Type: &schema.DataRef{Pos: typeParam.Pos, Name: typeParam.Name},
					}
				}
			}
		}

		if field == nil {
			queryMap[key] = value
			continue
		}

		switch field.Type.(type) {
		case *schema.Bytes, *schema.Map, *schema.Optional, *schema.Time,
			*schema.Unit, *schema.DataRef, *schema.Any:

		case *schema.Int, *schema.Float, *schema.String, *schema.Bool:
			if len(value) > 1 {
				return nil, fmt.Errorf("multiple values for %q are not supported", key)
			}
			if hasInvalidQueryChars(value[0]) {
				return nil, fmt.Errorf("complex value %q is not supported, use '@json=' instead", value[0])
			}
			queryMap[key] = value[0]

		case *schema.Array:
			for _, v := range value {
				if hasInvalidQueryChars(v) {
					return nil, fmt.Errorf("complex value %q is not supported, use '@json=' instead", v)
				}
			}
			queryMap[key] = value

		default:
			panic(fmt.Sprintf("unsupported type %T for query parameter field %q", field.Type, key))
		}
	}

	return queryMap, nil
}

func decodeQueryJSON(query string) (map[string]any, error) {
	decodedJSONStr, err := url.QueryUnescape(query)
	if err != nil {
		return nil, fmt.Errorf("failed to decode '@json' query parameter: %w", err)
	}

	// Unmarshal the JSON string into a map
	var resultMap map[string]any
	err = json.Unmarshal([]byte(decodedJSONStr), &resultMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse '@json' query parameter: %w", err)
	}

	return resultMap, nil
}

func hasInvalidQueryChars(s string) bool {
	return strings.ContainsAny(s, "{}[]|\\^`")
}
