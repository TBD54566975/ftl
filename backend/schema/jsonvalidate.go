package schema

import (
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type path []string

func (p path) String() string {
	return strings.TrimLeft(strings.Join(p, ""), ".")
}

// ValidateJSONValue validates a given JSON value against the provided schema.
func ValidateJSONValue(fieldType Type, path path, value any, sch *Schema) error { //nolint:maintidx
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
			if err := ValidateJSONValue(elementType, elemPath, elem, sch); err != nil {
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
			if err := ValidateJSONValue(keyType, elemPath, key.Interface(), sch); err != nil {
				return err
			}
			if err := ValidateJSONValue(valueType, elemPath, elem, sch); err != nil {
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
				if err := ValidateRequestMap(fieldType, path, valueMap, sch); err != nil {
					return err
				}
				typeMatches = true
			}
		case *TypeAlias:
			return ValidateJSONValue(d.Type, path, value, sch)
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
					if valueInt, ok := value.(int); ok {
						if t.Value == valueInt {
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
							return ValidateJSONValue(t.Value, path, vValue, sch)
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
			return ValidateJSONValue(fieldType.Type, path, value, sch)
		}
	}

	if !typeMatches {
		return fmt.Errorf("%s has wrong type, expected %s found %T", path, fieldType, value)
	}
	return nil
}

func ValidateRequestMap(ref *Ref, path path, request map[string]any, sch *Schema) error {
	data, err := sch.ResolveMonomorphised(ref)
	if err != nil {
		return err
	}

	var errs []error
	for _, field := range data.Fields {
		fieldPath := append(path, "."+field.Name) //nolint:gocritic

		value, haveValue := request[field.Name]
		if !haveValue && !allowMissingField(field) {
			errs = append(errs, fmt.Errorf("%s is required", fieldPath))
			continue
		}

		if haveValue {
			err := ValidateJSONValue(field.Type, fieldPath, value, sch)
			if err != nil {
				errs = append(errs, err)
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
