package ingress

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/schema"
)

// BuildRequestBody extracts the HttpRequest body from an HTTP request.
func BuildRequestBody(route *dal.IngressRoute, r *http.Request, sch *schema.Schema) ([]byte, error) {
	verb := &schema.Verb{}
	err := sch.ResolveToType(&schema.Ref{Name: route.Verb, Module: route.Module}, verb)
	if err != nil {
		return nil, err
	}

	request, ok := verb.Request.(*schema.Ref)
	if !ok {
		return nil, fmt.Errorf("verb %s input must be a data structure", verb.Name)
	}

	var body []byte

	var requestMap map[string]any

	if metadata, ok := verb.GetMetadataIngress().Get(); ok && metadata.Type == "http" {
		pathParameters := map[string]any{}
		matchSegments(route.Path, r.URL.Path, func(segment, value string) {
			pathParameters[segment] = value
		})

		httpRequestBody, err := extractHTTPRequestBody(route, r, request, sch)
		if err != nil {
			return nil, err
		}

		// Since the query and header parameters are a `map[string][]string`
		// we need to convert them before they go through the `transformFromAliasedFields` call
		// otherwise they will fail the type check.
		queryMap := make(map[string]any)
		for key, values := range r.URL.Query() {
			valuesAny := make([]any, len(values))
			for i, v := range values {
				valuesAny[i] = v
			}
			queryMap[key] = valuesAny
		}

		headerMap := make(map[string]any)
		for key, values := range r.Header {
			valuesAny := make([]any, len(values))
			for i, v := range values {
				valuesAny[i] = v
			}
			headerMap[key] = valuesAny
		}

		requestMap = map[string]any{}
		requestMap["method"] = r.Method
		requestMap["path"] = r.URL.Path
		requestMap["pathParameters"] = pathParameters
		requestMap["query"] = queryMap
		requestMap["headers"] = headerMap
		requestMap["body"] = httpRequestBody
	} else {
		var err error
		requestMap, err = buildRequestMap(route, r, request, sch)
		if err != nil {
			return nil, err
		}
	}

	requestMap, err = schema.TransformFromAliasedFields(request, sch, requestMap)
	if err != nil {
		return nil, err
	}

	err = schema.ValidateRequestMap(request, []string{request.String()}, requestMap, sch)
	if err != nil {
		return nil, err
	}

	body, err = json.Marshal(requestMap)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func extractHTTPRequestBody(route *dal.IngressRoute, r *http.Request, ref *schema.Ref, sch *schema.Schema) (any, error) {
	bodyField, err := getBodyField(ref, sch)
	if err != nil {
		return nil, err
	}

	if ref, ok := bodyField.Type.(*schema.Ref); ok {
		if err := sch.ResolveToType(ref, &schema.Data{}); err == nil {
			return buildRequestMap(route, r, ref, sch)
		}
	}

	bodyData, err := readRequestBody(r)
	if err != nil {
		return nil, err
	}

	return valueForData(bodyField.Type, bodyData)
}

func valueForData(typ schema.Type, data []byte) (any, error) {
	switch typ.(type) {
	case *schema.Ref:
		var bodyMap map[string]any
		err := json.Unmarshal(data, &bodyMap)
		if err != nil {
			return nil, fmt.Errorf("HTTP request body is not valid JSON: %w", err)
		}
		return bodyMap, nil

	case *schema.Array:
		var rawData []json.RawMessage
		err := json.Unmarshal(data, &rawData)
		if err != nil {
			return nil, fmt.Errorf("HTTP request body is not a valid JSON array: %w", err)
		}

		arrayData := make([]any, len(rawData))
		for i, rawElement := range rawData {
			var parsedElement any
			err := json.Unmarshal(rawElement, &parsedElement)
			if err != nil {
				return nil, fmt.Errorf("failed to parse array element: %w", err)
			}
			arrayData[i] = parsedElement
		}

		return arrayData, nil

	case *schema.Map:
		var bodyMap map[string]any
		err := json.Unmarshal(data, &bodyMap)
		if err != nil {
			return nil, fmt.Errorf("HTTP request body is not valid JSON: %w", err)
		}
		return bodyMap, nil

	case *schema.Bytes:
		return data, nil

	case *schema.String:
		return string(data), nil

	case *schema.Int:
		intVal, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse integer from request body: %w", err)
		}
		return intVal, nil

	case *schema.Float:
		floatVal, err := strconv.ParseFloat(string(data), 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse float from request body: %w", err)
		}
		return floatVal, nil

	case *schema.Bool:
		boolVal, err := strconv.ParseBool(string(data))
		if err != nil {
			return nil, fmt.Errorf("failed to parse boolean from request body: %w", err)
		}
		return boolVal, nil

	case *schema.Unit:
		return map[string]any{}, nil

	default:
		return nil, fmt.Errorf("unsupported data type %T", typ)
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

func buildRequestMap(route *dal.IngressRoute, r *http.Request, ref *schema.Ref, sch *schema.Schema) (map[string]any, error) {
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
		data, err := sch.ResolveMonomorphised(ref)
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
			if jsonAlias, ok := f.Alias(schema.AliasKindJSON).Get(); (ok && jsonAlias == key) || f.Name == key {
				field = f
			}
			for _, typeParam := range data.TypeParameters {
				if typeParam.Name == key {
					field = &schema.Field{
						Name: key,
						Type: &schema.Ref{Pos: typeParam.Pos, Name: typeParam.Name},
					}
				}
			}
		}

		if field == nil {
			queryMap[key] = value
			continue
		}

		val, err := valueForField(field.Type, value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse query parameter %q: %w", key, err)
		}

		if v, ok := val.Get(); ok {
			queryMap[key] = v
		}
	}

	return queryMap, nil
}

func valueForField(typ schema.Type, value []string) (optional.Option[any], error) {
	switch t := typ.(type) {
	case *schema.Bytes, *schema.Map, *schema.Time,
		*schema.Unit, *schema.Ref, *schema.Any:

	case *schema.Int, *schema.Float, *schema.String, *schema.Bool:
		if len(value) > 1 {
			return optional.None[any](), fmt.Errorf("multiple values are not supported")
		}
		if hasInvalidQueryChars(value[0]) {
			return optional.None[any](), fmt.Errorf("complex value %q is not supported, use '@json=' instead", value[0])
		}
		return optional.Some[any](value[0]), nil

	case *schema.Array:
		for _, v := range value {
			if hasInvalidQueryChars(v) {
				return optional.None[any](), fmt.Errorf("complex value %q is not supported, use '@json=' instead", v)
			}
		}
		return optional.Some[any](value), nil

	case *schema.Optional:
		if len(value) > 0 {
			return valueForField(t.Type, value)
		}

	default:
		panic(fmt.Sprintf("unsupported type %T", typ))
	}

	return optional.Some[any](value), nil
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
