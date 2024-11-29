package provisioner_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	proto "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
	"github.com/TBD54566975/ftl/backend/provisioner"
)

func TestSubGraphWithDirectDependencies(t *testing.T) {
	t.Run("returns a subgraph with direct dependencies", func(t *testing.T) {
		graph := &provisioner.ResourceGraph{}
		a := graph.AddNode(&proto.Resource{ResourceId: "a", Resource: &proto.Resource_Mysql{}})
		b := graph.AddNode(&proto.Resource{ResourceId: "b", Resource: &proto.Resource_Mysql{}})
		c := graph.AddNode(&proto.Resource{ResourceId: "c", Resource: &proto.Resource_Mysql{}})
		_ = graph.AddNode(&proto.Resource{ResourceId: "d", Resource: &proto.Resource_Mysql{}})

		graph.AddEdge(a, b)
		graph.AddEdge(b, c)

		subgraph := graph.WithDirectDependencies([]*proto.Resource{a})
		assert.Equal(t, []*proto.Resource{b, a}, subgraph.Resources())
	})
}
