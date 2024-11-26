package provisioner

import (
	"testing"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
)

func TestResourceEqual(t *testing.T) {
	t.Run("equal messages are equal", func(t *testing.T) {
		msg1 := &provisioner.Resource{
			ResourceId: "test",
			Resource: &provisioner.Resource_Postgres{
				Postgres: &provisioner.PostgresResource{},
			},
		}
		msg2 := &provisioner.Resource{
			ResourceId: "test",
			Resource: &provisioner.Resource_Postgres{
				Postgres: &provisioner.PostgresResource{},
			},
		}
		if !resourceEqual(msg1, msg2) {
			t.Errorf("expected messages to be equal, but they are not")
		}
	})
	t.Run("different resource types are not equal", func(t *testing.T) {
		msg1 := &provisioner.Resource{
			ResourceId: "test",
			Resource: &provisioner.Resource_Postgres{
				Postgres: &provisioner.PostgresResource{},
			},
		}
		msg2 := &provisioner.Resource{
			ResourceId: "test",
			Resource: &provisioner.Resource_Mysql{
				Mysql: &provisioner.MysqlResource{},
			},
		}
		if resourceEqual(msg1, msg2) {
			t.Errorf("expected messages to be different, but they are equal")
		}
	})
	t.Run("different outputs are still equal", func(t *testing.T) {
		msg1 := &provisioner.Resource{
			ResourceId: "test",
			Resource: &provisioner.Resource_Postgres{
				Postgres: &provisioner.PostgresResource{},
			},
		}
		msg2 := &provisioner.Resource{
			ResourceId: "test",
			Resource: &provisioner.Resource_Postgres{
				Postgres: &provisioner.PostgresResource{
					Output: &schemapb.DatabaseRuntime{
						Value: &schemapb.DatabaseRuntime_DsnDatabaseRuntime{
							DsnDatabaseRuntime: &schemapb.DSNDatabaseRuntime{
								Dsn: "foo",
							},
						},
					},
				},
			},
		}
		if !resourceEqual(msg1, msg2) {
			t.Errorf("expected messages to be equal, but they are not")
		}
	})
}
