package ingress

import (
	"fmt"

	"github.com/TBD54566975/ftl/backend/schema"
)

func transformAliasedFields(sch *schema.Schema, t schema.Type, obj any, aliaser func(obj map[string]any, field *schema.Field) string) error {
	if obj == nil {
		return nil
	}
	switch t := t.(type) {
	case *schema.DataRef:
		data, err := sch.ResolveDataRefMonomorphised(t)
		if err != nil {
			return fmt.Errorf("%s: failed to resolve data type: %w", t.Pos, err)
		}
		m, ok := obj.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: expected map, got %T", t.Pos, obj)
		}
		for _, field := range data.Fields {
			name := aliaser(m, field)
			if err := transformAliasedFields(sch, field.Type, m[name], aliaser); err != nil {
				return err
			}
		}

	case *schema.Array:
		a, ok := obj.([]any)
		if !ok {
			return fmt.Errorf("%s: expected array, got %T", t.Pos, obj)
		}
		for _, elem := range a {
			if err := transformAliasedFields(sch, t.Element, elem, aliaser); err != nil {
				return err
			}
		}

	case *schema.Map:
		m, ok := obj.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: expected map, got %T", t.Pos, obj)
		}
		for key, value := range m {
			if err := transformAliasedFields(sch, t.Key, key, aliaser); err != nil {
				return err
			}
			if err := transformAliasedFields(sch, t.Value, value, aliaser); err != nil {
				return err
			}
		}

	case *schema.Optional:
		if obj == nil {
			return nil
		}
		return transformAliasedFields(sch, t.Type, obj, aliaser)

	case *schema.Any, *schema.Bool, *schema.Bytes, *schema.Float, *schema.Int,
		*schema.String, *schema.Time, *schema.Unit:
	}
	return nil
}

func transformFromAliasedFields(dataRef *schema.DataRef, sch *schema.Schema, request map[string]any) (map[string]any, error) {
	return request, transformAliasedFields(sch, dataRef, request, func(obj map[string]any, field *schema.Field) string {
		if _, ok := obj[field.Name]; !ok && field.JSONAlias != "" && obj[field.JSONAlias] != nil {
			obj[field.Name] = obj[field.JSONAlias]
			delete(obj, field.JSONAlias)
		}
		return field.Name
	})
}

func transformToAliasedFields(dataRef *schema.DataRef, sch *schema.Schema, request map[string]any) (map[string]any, error) {
	return request, transformAliasedFields(sch, dataRef, request, func(obj map[string]any, field *schema.Field) string {
		if field.JSONAlias != "" && field.Name != field.JSONAlias {
			obj[field.JSONAlias] = obj[field.Name]
			delete(obj, field.Name)
			return field.JSONAlias
		}
		return field.Name
	})
}
