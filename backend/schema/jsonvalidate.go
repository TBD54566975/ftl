package schema

import (
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	sets "github.com/deckarep/golang-set/v2"
)

type path []string

func (p path) String() string {
	return strings.TrimLeft(strings.Join(p, ""), ".")
}

type encodingOptions struct {
	lenient bool
}

type EncodingOption func(option *encodingOptions)

func LenientMode() EncodingOption {
	return func(eo *encodingOptions) {
		eo.lenient = true
	}
}

// ValidateJSONValue validates a given JSON value against the provided schema.
func ValidateJSONValue(fieldType Type, path path, value any, sch *Schema, opts ...EncodingOption) error {
	cfg := &encodingOptions{}
	for _, opt := range opts {
		opt(cfg)
	}
	return validateJSONValue(fieldType, path, value, sch, cfg)
}

func validateJSONValue(fieldType Type, path path, value any, sch *Schema, opts *encodingOptions) error { //nolint:maintidx
	var typeMatches bool
	switch fieldType := fieldType.(type) {
	case *Any:
		typeMatches = true

	case *Unit:
		// TODO: Use type assertions consistently in this function rather than reflection.
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Map || rv.Len() != 0 {
			return fmt.Errorf("%s must be an empty map", path)
		}
		return nil

	case *Time:
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("time %s must be an RFC3339 formatted string", path)
		}
		_, err := time.Parse(time.RFC3339Nano, str)
		if err != nil {
			return fmt.Errorf("time %s must be an RFC3339 formatted string: %w", path, err)
		}
		return nil

	case *Int:
		switch value := value.(type) {
		case int64, float64:
			typeMatches = true
		case string:
			if _, err := strconv.ParseInt(value, 10, 64); err == nil {
				typeMatches = true
			}
		}

	case *Float:
		switch value := value.(type) {
		case float64:
			typeMatches = true
		case string:
			if _, err := strconv.ParseFloat(value, 64); err == nil {
				typeMatches = true
			}
		}

	case *String:
		_, typeMatches = value.(string)

	case *Bool:
		switch value := value.(type) {
		case bool:
			typeMatches = true
		case string:
			if _, err := strconv.ParseBool(value); err == nil {
				typeMatches = true
			}
		}

	case *Array:
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Slice {
			return fmt.Errorf("%s is not a slice", path)
		}
		elementType := fieldType.Element
		for i := range rv.Len() {
			elemPath := append(path, fmt.Sprintf("[%d]", i)) //nolint:gocritic
			elem := rv.Index(i).Interface()
			if err := validateJSONValue(elementType, elemPath, elem, sch, opts); err != nil {
				return err
			}
		}
		typeMatches = true

	case *Map:
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Map {
			return fmt.Errorf("%s is not a map", path)
		}
		keyType := fieldType.Key
		valueType := fieldType.Value
		for _, key := range rv.MapKeys() {
			elemPath := append(path, fmt.Sprintf("[%q]", key)) //nolint:gocritic
			elem := rv.MapIndex(key).Interface()
			if err := validateJSONValue(keyType, elemPath, key.Interface(), sch, opts); err != nil {
				return err
			}
			if err := validateJSONValue(valueType, elemPath, elem, sch, opts); err != nil {
				return err
			}
		}
		typeMatches = true
	case *Ref:
		decl, ok := sch.Resolve(fieldType).Get()
		if !ok {
			return fmt.Errorf("unknown ref %v", fieldType)
		}

		switch d := decl.(type) {
		case *Data:
			if valueMap, ok := value.(map[string]any); ok {
				transformedMap, err := TransformFromAliasedFields(fieldType, sch, valueMap)
				if err != nil {
					return fmt.Errorf("failed to transform aliased fields: %w", err)
				}

				if err := validateRequestMap(fieldType, path, transformedMap, sch, opts); err != nil {
					return err
				}
				typeMatches = true
			}
		case *TypeAlias:
			return validateJSONValue(d.Type, path, value, sch, opts)
		case *Enum:
			var inputName any
			inputName = value
			for _, v := range d.Variants {
				switch t := v.Value.(type) {
				case *StringValue:
					if valueStr, ok := value.(string); ok {
						if t.Value == valueStr {
							typeMatches = true
							break
						}
					}
				case *IntValue:
					switch value := value.(type) {
					case int, int64:
						if t.Value == value {
							typeMatches = true
							break
						}
					case float64:
						if float64(t.Value) == value {
							typeMatches = true
							break
						}
					}
				case *TypeValue:
					if reqVariant, ok := value.(map[string]any); ok {
						vName, ok := reqVariant["name"]
						if !ok {
							return fmt.Errorf(`missing name field in enum type %q: expected structure is `+
								"{\"name\": \"<variant name>\", \"value\": <variant value>}", value)
						}
						vNameStr, ok := vName.(string)
						if !ok {
							return fmt.Errorf(`invalid type for enum %q; name field must be a string, was %T`,
								fieldType, vName)
						}
						inputName = fmt.Sprintf("%q", vNameStr)

						vValue, ok := reqVariant["value"]
						if !ok {
							return fmt.Errorf(`missing value field in enum type %q: expected structure is `+
								"{\"name\": \"<variant name>\", \"value\": <variant value>}", value)
						}

						if v.Name == vNameStr {
							return validateJSONValue(t.Value, path, vValue, sch, opts)
						}
					} else {
						return fmt.Errorf(`malformed enum type %s: expected structure is `+
							"{\"name\": \"<variant name>\", \"value\": <variant value>}", path)
					}
				}
			}
			if !typeMatches {
				return fmt.Errorf("%s is not a valid variant of enum %s", inputName, fieldType)
			}

		case *Config, *Database, *Secret, *Verb, *FSM, *Topic, *Subscription:

		}

	case *Bytes:
		_, typeMatches = value.([]byte)
		if bodyStr, ok := value.(string); ok {
			_, err := base64.StdEncoding.DecodeString(bodyStr)
			if err != nil {
				return fmt.Errorf("%s is not a valid base64 string", path)
			}
			typeMatches = true
		}

	case *Optional:
		if value == nil {
			typeMatches = true
		} else {
			return validateJSONValue(fieldType.Type, path, value, sch, opts)
		}
	}

	if !typeMatches {
		return fmt.Errorf("%s has wrong type, expected %s found %T", path, fieldType, value)
	}
	return nil
}

// ValidateRequestMap validates a given JSON map against the provided schema.
func ValidateRequestMap(ref *Ref, path path, request map[string]any, sch *Schema, opts ...EncodingOption) error {
	cfg := &encodingOptions{}
	for _, opt := range opts {
		opt(cfg)
	}
	return validateRequestMap(ref, path, request, sch, cfg)
}

// ValidateRequestMap validates a given JSON map against the provided schema.
func validateRequestMap(ref *Ref, path path, request map[string]any, sch *Schema, opts *encodingOptions) error {
	symbol, err := sch.ResolveRequestResponseType(ref)
	if err != nil {
		return err
	}

	var errs []error
	if data, ok := symbol.(*Data); ok {
		validFields := sets.NewSet[string]()
		for _, field := range data.Fields {
			validFields.Add(field.Name)
			fieldPath := append(path, "."+field.Name) //nolint:gocritic

			value, haveValue := request[field.Name]
			if !haveValue && !allowMissingField(field) {
				errs = append(errs, fmt.Errorf("%s is required", fieldPath))
				continue
			}

			if haveValue {
				err := validateJSONValue(field.Type, fieldPath, value, sch, opts)
				if err != nil {
					errs = append(errs, err)
				}
			}
		}

		if !opts.lenient {
			for key := range request {
				if !validFields.Contains(key) {
					errs = append(errs, fmt.Errorf("%s is not a valid field", append(path, "."+key)))
				}
			}
		}
	}

	return errors.Join(errs...)
}

// Fields of these types can be omitted from the JSON representation.
func allowMissingField(field *Field) bool {
	switch field.Type.(type) {
	case *Optional, *Any, *Array, *Map, *Bytes, *Unit:
		return true

	case *Bool, *Ref, *Float, *Int, *String, *Time:
	}
	return false
}

func TransformAliasedFields(sch *Schema, t Type, obj any, aliaser func(obj map[string]any, field *Field) string) error {
	if obj == nil {
		return nil
	}
	switch t := t.(type) {
	case *Ref:
		decl, ok := sch.Resolve(t).Get()
		if !ok {
			return fmt.Errorf("%s: failed to resolve ref %s", t.Pos, t)
		}
		switch decl := decl.(type) {
		case *Data:
			data, err := sch.ResolveMonomorphised(t)
			if err != nil {
				return fmt.Errorf("%s: failed to resolve data type: %w", t.Pos, err)
			}
			m, ok := obj.(map[string]any)
			if !ok {
				return fmt.Errorf("%s: expected map, got %T", t.Pos, obj)
			}
			for _, field := range data.Fields {
				name := aliaser(m, field)
				if err := TransformAliasedFields(sch, field.Type, m[name], aliaser); err != nil {
					return err
				}
			}
		case *Enum:
			if decl.IsValueEnum() {
				return nil
			}

			// type enum
			m, ok := obj.(map[string]any)
			if !ok {
				return fmt.Errorf("%s: expected map, got %T", t.Pos, obj)
			}
			name, ok := m["name"]
			if !ok {
				return fmt.Errorf("%s: expected type enum request to have 'name' field", t.Pos)
			}
			nameStr, ok := name.(string)
			if !ok {
				return fmt.Errorf("%s: expected 'name' field to be a string, got %T", t.Pos, name)
			}

			value, ok := m["value"]
			if !ok {
				return fmt.Errorf("%s: expected type enum request to have 'value' field", t.Pos)
			}

			for _, v := range decl.Variants {
				if v.Name == nameStr {
					if err := TransformAliasedFields(sch, v.Value.(*TypeValue).Value, value, aliaser); err != nil { //nolint:forcetypeassert
						return err
					}
				}
			}
		case *TypeAlias:
			return TransformAliasedFields(sch, decl.Type, obj, aliaser)
		case *Config, *Database, *FSM, *Secret, *Verb, *Topic, *Subscription:
			return fmt.Errorf("%s: unsupported ref type %T", t.Pos, decl)
		}

	case *Array:
		a, ok := obj.([]any)
		if !ok {
			return fmt.Errorf("%s: expected array, got %T", t.Pos, obj)
		}
		for _, elem := range a {
			if err := TransformAliasedFields(sch, t.Element, elem, aliaser); err != nil {
				return err
			}
		}

	case *Map:
		m, ok := obj.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: expected map, got %T", t.Pos, obj)
		}
		for key, value := range m {
			if err := TransformAliasedFields(sch, t.Key, key, aliaser); err != nil {
				return err
			}
			if err := TransformAliasedFields(sch, t.Value, value, aliaser); err != nil {
				return err
			}
		}

	case *Optional:
		return TransformAliasedFields(sch, t.Type, obj, aliaser)

	case *Any, *Bool, *Bytes, *Float, *Int,
		*String, *Time, *Unit:
	}
	return nil
}

func TransformFromAliasedFields(ref *Ref, sch *Schema, request map[string]any) (map[string]any, error) {
	return request, TransformAliasedFields(sch, ref, request, func(obj map[string]any, field *Field) string {
		if jsonAlias, ok := field.Alias(AliasKindJSON).Get(); ok {
			if _, ok := obj[field.Name]; !ok && obj[jsonAlias] != nil {
				obj[field.Name] = obj[jsonAlias]
				delete(obj, jsonAlias)
			}
		}
		return field.Name
	})
}

func TransformToAliasedFields(ref *Ref, sch *Schema, request map[string]any) (map[string]any, error) {
	return request, TransformAliasedFields(sch, ref, request, func(obj map[string]any, field *Field) string {
		if jsonAlias, ok := field.Alias(AliasKindJSON).Get(); ok && field.Name != jsonAlias {
			obj[jsonAlias] = obj[field.Name]
			delete(obj, field.Name)
			return jsonAlias
		}
		return field.Name
	})
}
