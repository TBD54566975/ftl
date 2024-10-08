package ingress

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/TBD54566975/ftl/backend/controller/dal/model"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/slices"
)

func GetIngressRoute(routes []model.IngressRoute, method string, path string) (*model.IngressRoute, error) {
	var matchedRoutes = slices.Filter(routes, func(route model.IngressRoute) bool {
		return matchSegments(route.Path, path, func(segment, value string) {})
	})

	if len(matchedRoutes) == 0 {
		return nil, libdal.ErrNotFound
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

func getField(name string, ref *schema.Ref, sch *schema.Schema) (*schema.Field, error) {
	data, err := sch.ResolveMonomorphised(ref)
	if err != nil {
		return nil, err
	}
	var bodyField *schema.Field
	for _, field := range data.Fields {
		if field.Name == name {
			bodyField = field
			break
		}
	}

	if bodyField == nil {
		return nil, fmt.Errorf("verb %s must have a %q field", ref.Name, name)
	}

	return bodyField, nil
}
