package schema

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

type ProvisionedResource struct {
	// Kind is the kind of resource provisioned.
	Kind string
	// ID of the provisioned resource.
	ID string
	// Config is the subset of the schema element's configuration that is used to create the resource.
	// changes to this config are used to check if the resource needs to be updated.
	Config any
}

// Provisioned is a schema element that provisioner acts on to create a runtime resource.
type Provisioned interface {
	// Returns the resources provisioned from this schema element.
	GetProvisioned() []*ProvisionedResource
}
