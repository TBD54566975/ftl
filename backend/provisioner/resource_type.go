package provisioner

import (
	provisioner "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
	"github.com/TBD54566975/ftl/internal/schema"
)

// TypeOf returns the resource type of the given resource
func TypeOf(r *provisioner.Resource) schema.ResourceType {
	switch r.Resource.(type) {
	case *provisioner.Resource_Module:
		return schema.ResourceTypeModule
	case *provisioner.Resource_Mysql:
		return schema.ResourceTypeMysql
	case *provisioner.Resource_Postgres:
		return schema.ResourceTypePostgres
	case *provisioner.Resource_SqlMigration:
		return schema.ResourceTypeSQLMigration
	case *provisioner.Resource_Topic:
		return schema.ResourceTypeTopic
	case *provisioner.Resource_Subscription:
		return schema.ResourceTypeSubscription
	case *provisioner.Resource_Runner:
		return schema.ResourceTypeRunner
	default:
		return schema.ResourceTypeUnknown
	}
}
