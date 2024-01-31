package ingress

import (
	"github.com/TBD54566975/ftl/backend/schema"
)

func transformFromAliasedFields(dataRef *schema.DataRef, sch *schema.Schema, request map[string]any) (map[string]any, error) {
	data, err := sch.ResolveDataRefMonomorphised(dataRef)
	if err != nil {
		return nil, err
	}

	for _, field := range data.Fields {
		if _, ok := request[field.Name]; !ok && field.Alias != "" && request[field.Alias] != nil {
			request[field.Name] = request[field.Alias]
			delete(request, field.Alias)
		}

		if d, ok := field.Type.(*schema.DataRef); ok {
			if _, found := request[field.Name]; found {
				rMap, err := transformFromAliasedFields(d, sch, request[field.Name].(map[string]any))
				if err != nil {
					return nil, err
				}
				request[field.Name] = rMap
			}
		}
	}

	return request, nil
}

func transformToAliasedFields(dataRef *schema.DataRef, sch *schema.Schema, request map[string]any) (map[string]any, error) {
	data, err := sch.ResolveDataRefMonomorphised(dataRef)
	if err != nil {
		return nil, err
	}

	for _, field := range data.Fields {
		if field.Alias != "" && field.Name != field.Alias {
			request[field.Alias] = request[field.Name]
			delete(request, field.Name)
		}

		if d, ok := field.Type.(*schema.DataRef); ok {
			if _, found := request[field.Name]; found {
				rMap, err := transformToAliasedFields(d, sch, request[field.Name].(map[string]any))
				if err != nil {
					return nil, err
				}
				request[field.Name] = rMap
			}
		}
	}

	return request, nil
}
