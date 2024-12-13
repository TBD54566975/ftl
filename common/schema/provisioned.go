package schema

import (
	"reflect"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/TBD54566975/ftl/common/slices"
)

// ResourceType is a type of resource used to configure provisioners
type ResourceType string

const (
	ResourceTypeUnknown      ResourceType = "unknown"
	ResourceTypePostgres     ResourceType = "postgres"
	ResourceTypeMysql        ResourceType = "mysql"
	ResourceTypeModule       ResourceType = "module"
	ResourceTypeSQLMigration ResourceType = "sql-migration"
	ResourceTypeTopic        ResourceType = "topic"
	ResourceTypeSubscription ResourceType = "subscription"
	ResourceTypeRunner       ResourceType = "runner"
)

// ProvisionedResource is metadata of a resource that can be provisioned
// from a schema element.
//
// One schema element can result into multiple different resources.
// For example, a database node can be provisioned into a physical database, and a schema migration
type ProvisionedResource struct {
	// Kind is the kind of resource provisioned.
	Kind ResourceType
	// Config is the subset of the schema element's configuration that is used to create the resource.
	// changes to this config are used to check if the resource needs to be updated.
	Config any
}

func (r *ProvisionedResource) IsEqual(other *ProvisionedResource) bool {
	return cmp.Equal(r, other)
}

// Provisioned is a schema element that provisioner acts on to create a runtime resource.
//
//sumtype:decl
type Provisioned interface {
	Node
	// Returns the resources provisioned from this schema element.
	GetProvisioned() ResourceSet
	ResourceID() string

	provisioned()
}

type ResourceSet []*ProvisionedResource

func (s ResourceSet) IsEqual(other ResourceSet) bool {
	return cmp.Equal(s, other, cmpopts.SortSlices(func(x, y *ProvisionedResource) bool {
		return x.Kind < y.Kind
	}))
}

func (s ResourceSet) FilterByType(kinds ...ResourceType) ResourceSet {
	return slices.Filter(s, func(x *ProvisionedResource) bool {
		for _, k := range kinds {
			if x.Kind == k {
				return true
			}
		}
		return false
	})
}

func (s ResourceSet) Get(kind ResourceType) *ProvisionedResource {
	result, _ := slices.Find(s, func(x *ProvisionedResource) bool {
		return x.Kind == kind
	})
	return result
}

func GetProvisionedResources(n Node) ResourceSet {
	if n == nil || reflect.ValueOf(n).IsNil() {
		return ResourceSet{}
	}

	var resources []*ProvisionedResource
	Visit(n, func(n Node, next func() error) error { //nolint:errcheck
		if p, ok := n.(Provisioned); ok {
			resources = append(resources, p.GetProvisioned()...)
		}
		return next()
	})
	return resources
}

func GetProvisioned(root Node) map[string]Provisioned {
	if root == nil || reflect.ValueOf(root).IsNil() {
		return map[string]Provisioned{}
	}

	result := map[string]Provisioned{}
	Visit(root, func(n Node, next func() error) error { //nolint:errcheck
		if p, ok := n.(Provisioned); ok {
			result[p.ResourceID()] = p
		}
		return next()
	})
	return result
}
