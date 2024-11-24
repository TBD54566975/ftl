package provisioner

import (
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
)

// ResourceType is a type of resource used to configure provisioners
type ResourceType string

const (
	ResourceTypeUnknown      ResourceType = "unknown"
	ResourceTypePostgres     ResourceType = "postgres"
	ResourceTypeMysql        ResourceType = "mysql"
	ResourceTypeModule       ResourceType = "module"
	ResourceTypeSQLMigration ResourceType = "sql-migration"
)

// TypeOf returns the resource type of the given resource
func TypeOf(r *provisioner.Resource) ResourceType {
	switch r.Resource.(type) {
	case *provisioner.Resource_Module:
		return ResourceTypeModule
	case *provisioner.Resource_Mysql:
		return ResourceTypeMysql
	case *provisioner.Resource_Postgres:
		return ResourceTypePostgres
	case *provisioner.Resource_SqlMigration:
		return ResourceTypeSQLMigration
	default:
		return ResourceTypeUnknown
	}
}
