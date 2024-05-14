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
	case *schema.Ref:
		switch decl := sch.ResolveRef(t).(type) {
		case *schema.Data:
			data, err := sch.ResolveRefMonomorphised(t)
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
		case *schema.Enum:
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
					if err := transformAliasedFields(sch, v.Value.(*schema.TypeValue).Value, value, aliaser); err != nil { //nolint:forcetypeassert
						return err
					}
				}
			}
		case *schema.Config, *schema.Database, *schema.FSM, *schema.Secret, *schema.Verb:
			return fmt.Errorf("%s: unsupported ref type %T", t.Pos, decl)
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
		return transformAliasedFields(sch, t.Type, obj, aliaser)

	case *schema.Any, *schema.Bool, *schema.Bytes, *schema.Float, *schema.Int,
		*schema.String, *schema.Time, *schema.Unit:
	}
	return nil
}

func transformFromAliasedFields(ref *schema.Ref, sch *schema.Schema, request map[string]any) (map[string]any, error) {
	return request, transformAliasedFields(sch, ref, request, func(obj map[string]any, field *schema.Field) string {
		if jsonAlias, ok := field.Alias(schema.AliasKindJSON).Get(); ok {
			if _, ok := obj[field.Name]; !ok && obj[jsonAlias] != nil {
				obj[field.Name] = obj[jsonAlias]
				delete(obj, jsonAlias)
			}
		}
		return field.Name
	})
}

func transformToAliasedFields(ref *schema.Ref, sch *schema.Schema, request map[string]any) (map[string]any, error) {
	return request, transformAliasedFields(sch, ref, request, func(obj map[string]any, field *schema.Field) string {
		if jsonAlias, ok := field.Alias(schema.AliasKindJSON).Get(); ok && field.Name != jsonAlias {
			obj[jsonAlias] = obj[field.Name]
			delete(obj, field.Name)
			return jsonAlias
		}
		return field.Name
	})
}
