package provisioner

import (
	"testing"

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
					Output: &provisioner.PostgresResource_PostgresResourceOutput{
						ReadDsn: "foo",
					},
				},
			},
		}
		if !resourceEqual(msg1, msg2) {
			t.Errorf("expected messages to be equal, but they are not")
		}
	})
}
