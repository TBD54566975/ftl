package ingress

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

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

func ValidateCallBody(body []byte, verbRef *schema.VerbRef, sch *schema.Schema) error {
	var requestMap map[string]any
	err := json.Unmarshal(body, &requestMap)
	if err != nil {
		return fmt.Errorf("HTTP request body is not valid JSON: %w", err)
	}

	verb := sch.ResolveVerbRef(verbRef)
	if verb == nil {
		return fmt.Errorf("unknown verb %s", verbRef)
	}

	dataRef := verb.Request

	return validateRequestMap(dataRef, []string{dataRef.String()}, requestMap, sch)
}

// ValidateAndExtractBody validates the request body against the schema and extracts the request body as a JSON blob.
func ValidateAndExtractBody(route *dal.IngressRoute, r *http.Request, sch *schema.Schema) ([]byte, error) {
	requestMap, err := buildRequestMap(route, r)
	if err != nil {
		return nil, err
	}

	verb := sch.ResolveVerbRef(&schema.VerbRef{Name: route.Verb, Module: route.Module})
	if verb == nil {
		return nil, fmt.Errorf("unknown verb %s", route.Verb)
	}

	dataRef := verb.Request
	if dataRef.Module == "" {
		dataRef.Module = route.Module
	}

	err = validateRequestMap(dataRef, []string{dataRef.String()}, requestMap, sch)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(requestMap)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func buildRequestMap(route *dal.IngressRoute, r *http.Request) (map[string]any, error) {
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
		queryMap, err := parseQueryParams(r.URL.Query())
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
	case *schema.Optional:
		if value == nil {
			typeMatches = true
		} else {
			return validateValue(fieldType.Type, path, value, sch)
		}

	default:
		return fmt.Errorf("%s has unsupported type %T", path, fieldType)
	}

	if !typeMatches {
		return fmt.Errorf("%s has wrong type, expected %s found %T", path, fieldType, value)
	}
	return nil
}

func parseQueryParams(values url.Values) (map[string]any, error) {
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
			return nil, fmt.Errorf("complex key '%s' is not supported, use '@json=' instead", key)
		}
		if len(value) == 1 {
			if hasInvalidQueryChars(value[0]) {
				return nil, fmt.Errorf("complex value '%s' is not supported, use '@json=' instead", value[0])
			}
			queryMap[key] = value[0]
		} else {
			for _, v := range value {
				if hasInvalidQueryChars(v) {
					return nil, fmt.Errorf("complex value '%s' is not supported, use '@json=' instead", v)
				}
			}
			// Assign as an array of strings if there are multiple values for the key
			queryMap[key] = value
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
