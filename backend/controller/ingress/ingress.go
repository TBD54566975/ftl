package ingress

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/TBD54566975/ftl/backend/common/slices"
	"github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/schema"
)

type path []string

func (p path) String() string {
	return strings.TrimLeft(strings.Join(p, ""), ".")
}

func GetIngressRoute(routes []dal.IngressRoute, method string, path string) (*dal.IngressRoute, error) {
	var matchedRoutes = slices.Filter(routes, func(route dal.IngressRoute) bool {
		return matchSegments(route.Path, path, func(segment, value string) {})
	})

	if len(matchedRoutes) == 0 {
		return nil, dal.ErrNotFound
	}

	// TODO: add load balancing at some point
	route := matchedRoutes[rand.Intn(len(matchedRoutes))] //nolint:gosec
	return &route, nil
}

func matchSegments(pattern, urlPath string, onMatch func(segment, value string)) bool {
	patternSegments := strings.Split(strings.Trim(pattern, "/"), "/")
	urlSegments := strings.Split(strings.Trim(urlPath, "/"), "/")

	if len(patternSegments) != len(urlSegments) {
		return false
	}

	for i, segment := range patternSegments {
		if segment == "" && urlSegments[i] == "" {
			continue // Skip empty segments
		}

		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			key := strings.Trim(segment, "{}") // Dynamic segment
			onMatch(key, urlSegments[i])
		} else if segment != urlSegments[i] {
			return false
		}
	}
	return true
}

func SetDefaultContentType(headers map[string][]string) {
	if _, hasContentType := headers["Content-Type"]; !hasContentType {
		headers["Content-Type"] = []string{"application/json"}
	}
}

func ResponseBodyForContentType(headers map[string][]string, body []byte) ([]byte, error) {
	if contentType, hasContentType := headers["Content-Type"]; hasContentType {
		if strings.HasPrefix(contentType[0], "text/") {
			var textContent string
			if err := json.Unmarshal(body, &textContent); err != nil {
				return nil, err
			}
			return []byte(textContent), nil
		}
	}

	return body, nil
}

func ValidateCallBody(body []byte, verbRef *schema.VerbRef, sch *schema.Schema) error {
	verb := sch.ResolveVerbRef(verbRef)
	if verb == nil {
		return fmt.Errorf("unknown verb %s", verbRef)
	}

	var requestMap map[string]any
	err := json.Unmarshal(body, &requestMap)
	if err != nil {
		return fmt.Errorf("HTTP request body is not valid JSON: %w", err)
	}

	return validateValue(verb.Request, []string{verb.Request.String()}, requestMap, sch)
}

// ValidateAndExtractRequestBody extracts the HttpRequest body from an HTTP request.
func ValidateAndExtractRequestBody(route *dal.IngressRoute, r *http.Request, sch *schema.Schema) ([]byte, error) {
	verb := sch.ResolveVerbRef(&schema.VerbRef{Name: route.Verb, Module: route.Module})
	if verb == nil {
		return nil, fmt.Errorf("unknown verb %s", route.Verb)
	}

	var request *schema.DataRef
	var body []byte
	if metadata, ok := verb.GetMetadataIngress().Get(); ok && metadata.Type == "http" {
		pathParameters := map[string]string{}
		matchSegments(route.Path, r.URL.Path, func(segment, value string) {
			pathParameters[segment] = value
		})

		request, ok = verb.Request.(*schema.DataRef)
		if !ok {
			return nil, fmt.Errorf("verb %s input must be a data structure", verb.Name)
		}

		bodyMap, err := buildRequest(route, r, request, sch)
		if err != nil {
			return nil, err
		}

		requestMap := map[string]any{}
		requestMap["method"] = r.Method
		requestMap["path"] = r.URL.Path
		requestMap["pathParameters"] = pathParameters
		requestMap["query"] = r.URL.Query()
		requestMap["headers"] = r.Header
		requestMap["body"] = bodyMap

		requestMap, err = transformAliasedFields(request, sch, requestMap)
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
	} else {
		request, ok = verb.Request.(*schema.DataRef)
		if !ok {
			return nil, fmt.Errorf("verb %s input must be a data structure", verb.Name)
		}

		requestMap, err := buildRequest(route, r, request, sch)
		if err != nil {
			return nil, err
		}

		err = validateRequestMap(request, []string{verb.Request.String()}, requestMap, sch)
		if err != nil {
			return nil, err
		}

		body, err = json.Marshal(requestMap)
		if err != nil {
			return nil, err
		}
	}

	return body, nil
}

func buildRequest(route *dal.IngressRoute, r *http.Request, dataRef *schema.DataRef, sch *schema.Schema) (map[string]any, error) {
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
		data := sch.ResolveDataRef(dataRef)
		if data == nil {
			return nil, fmt.Errorf("unknown data %v", dataRef)
		}

		if len(dataRef.TypeParameters) > 0 {
			var err error
			data, err = data.Monomorphise(dataRef.TypeParameters...)
			if err != nil {
				return nil, err
			}
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
	data := sch.ResolveDataRef(dataRef)
	if data == nil {
		return fmt.Errorf("unknown data %v", dataRef)
	}

	if len(dataRef.TypeParameters) > 0 {
		var err error
		data, err = data.Monomorphise(dataRef.TypeParameters...)
		if err != nil {
			return err
		}
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

func validateValue(fieldType schema.Type, path path, value any, sch *schema.Schema) error {
	var typeMatches bool
	switch fieldType := fieldType.(type) {
	case *schema.Any:
		typeMatches = true

	case *schema.Unit:
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Map || rv.Len() != 0 {
			return fmt.Errorf("%s must be an empty map", path)
		}
		return nil

	case *schema.Time:
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("time %s must be an RFC3339 formatted string", path)
		}
		_, err := time.Parse(time.RFC3339Nano, str)
		if err != nil {
			return fmt.Errorf("time %s must be an RFC3339 formatted string: %w", path, err)
		}
		return nil

	case *schema.Int:
		switch value := value.(type) {
		case float64:
			typeMatches = true
		case string:
			if _, err := strconv.ParseFloat(value, 64); err == nil {
				typeMatches = true
			}
		}

	case *schema.Float:
		switch value := value.(type) {
		case float64:
			typeMatches = true
		case string:
			if _, err := strconv.ParseFloat(value, 64); err == nil {
				typeMatches = true
			}
		}

	case *schema.String:
		_, typeMatches = value.(string)

	case *schema.Bool:
		switch value := value.(type) {
		case bool:
			typeMatches = true
		case string:
			if _, err := strconv.ParseBool(value); err == nil {
				typeMatches = true
			}
		}

	case *schema.Array:
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Slice {
			return fmt.Errorf("%s is not a slice", path)
		}
		elementType := fieldType.Element
		for i := 0; i < rv.Len(); i++ {
			elemPath := append(path, fmt.Sprintf("[%d]", i)) //nolint:gocritic
			elem := rv.Index(i).Interface()
			if err := validateValue(elementType, elemPath, elem, sch); err != nil {
				return err
			}
		}
		typeMatches = true

	case *schema.Map:
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Map {
			return fmt.Errorf("%s is not a map", path)
		}
		keyType := fieldType.Key
		valueType := fieldType.Value
		for _, key := range rv.MapKeys() {
			elemPath := append(path, fmt.Sprintf("[%q]", key)) //nolint:gocritic
			elem := rv.MapIndex(key).Interface()
			if err := validateValue(keyType, elemPath, key.Interface(), sch); err != nil {
				return err
			}
			if err := validateValue(valueType, elemPath, elem, sch); err != nil {
				return err
			}
		}
		typeMatches = true

	case *schema.DataRef:
		if valueMap, ok := value.(map[string]any); ok {
			if err := validateRequestMap(fieldType, path, valueMap, sch); err != nil {
				return err
			}
			typeMatches = true
		}

	case *schema.Bytes:
		_, typeMatches = value.([]byte)
		if bodyStr, ok := value.(string); ok {
			_, err := base64.StdEncoding.DecodeString(bodyStr)
			if err != nil {
				return fmt.Errorf("%s is not a valid base64 string", path)
			}
			typeMatches = true
		}

	case *schema.Optional:
		if value == nil {
			typeMatches = true
		} else {
			return validateValue(fieldType.Type, path, value, sch)
		}

	case *schema.TypeParameter:
		panic("data structures with type parameters should be monomorphised")
	}

	if !typeMatches {
		return fmt.Errorf("%s has wrong type, expected %s found %T", path, fieldType, value)
	}
	return nil
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
			if f.Name == key {
				field = f
			}
			for _, typeParam := range data.TypeParameters {
				if typeParam.String() == key {
					field = &schema.Field{
						Name: key,
						Type: typeParam,
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
			*schema.Unit, *schema.DataRef, *schema.Any, *schema.TypeParameter:

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

func transformAliasedFields(dataRef *schema.DataRef, sch *schema.Schema, request map[string]any) (map[string]any, error) {
	data := sch.ResolveDataRef(dataRef)
	if len(dataRef.TypeParameters) > 0 {
		var err error
		data, err = data.Monomorphise(dataRef.TypeParameters...)
		if err != nil {
			return nil, err
		}
	}

	for _, field := range data.Fields {
		if _, ok := request[field.Name]; !ok && field.Alias != "" {
			request[field.Name] = request[field.Alias]
			delete(request, field.Alias)
		}

		if d, ok := field.Type.(*schema.DataRef); ok {
			if _, found := request[field.Name]; found {
				rMap, err := transformAliasedFields(d, sch, request[field.Name].(map[string]any))
				if err != nil {
					return nil, err
				}
				request[field.Name] = rMap
			}
		}
	}

	return request, nil
}
