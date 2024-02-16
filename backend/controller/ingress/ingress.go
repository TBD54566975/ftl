package ingress

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/slices"
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

func getBodyField(dataRef *schema.DataRef, sch *schema.Schema) (*schema.Field, error) {
	data, err := sch.ResolveDataRefMonomorphised(dataRef)
	if err != nil {
		return nil, err
	}
	var bodyField *schema.Field
	for _, field := range data.Fields {
		if field.Name == "body" {
			bodyField = field
			break
		}
	}

	if bodyField == nil {
		return nil, fmt.Errorf("verb %s must have a 'body' field", dataRef.Name)
	}

	return bodyField, nil
}

func validateValue(fieldType schema.Type, path path, value any, sch *schema.Schema) error {
	var typeMatches bool
	switch fieldType := fieldType.(type) {
	case *schema.Any:
		typeMatches = true

	case *schema.Unit:
		// TODO: Use type assertions consistently in this function rather than reflection.
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
		case int64, float64:
			typeMatches = true
		case string:
			if _, err := strconv.ParseInt(value, 10, 64); err == nil {
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
	}

	if !typeMatches {
		return fmt.Errorf("%s has wrong type, expected %s found %T", path, fieldType, value)
	}
	return nil
}
