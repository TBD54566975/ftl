package infra_test

import (
	"testing"

	"github.com/TBD54566975/ftl/internal/infra"
	"gotest.tools/v3/assert"
)

func TestRGEdges(t *testing.T) {
	t.Run("returns incoming edges", func(t *testing.T) {
		g := infra.NewResourceGraph()
		deployment, _ := g.AddNode(&infra.Deployment{}, g.Root(), nil)
		assert.Equal(t, len(g.In(deployment)), 1)
	})
	t.Run("returns outgoing edges", func(t *testing.T) {
		g := infra.NewResourceGraph()
		_, _ = g.AddNode(&infra.Deployment{}, g.Root(), nil)
		_, _ = g.AddNode(&infra.Deployment{}, g.Root(), nil)
		assert.Equal(t, len(g.Out(g.Root())), 2)
	})
	t.Run("deleting last incoming edge removes the node", func(t *testing.T) {
		g := infra.NewResourceGraph()
		node, e1 := g.AddNode(infra.NewModule("A", "dep1"), g.Root(), nil)

		e2 := g.AddEdge(g.Root(), node, nil)
		assert.Assert(t, g.ById(node.Resource.Id()) != nil)

		g.DelEdge(e1)
		assert.Assert(t, g.ById(node.Resource.Id()) != nil)

		g.DelEdge(e2)
		assert.Assert(t, g.ById(node.Resource.Id()) == nil)
	})
}

func TestRGQueries(t *testing.T) {
	t.Run("queries by module", func(t *testing.T) {
		g := infra.NewResourceGraph()
		_, _ = g.AddNode(infra.NewModule("A", "dep1"), g.Root(), nil)
		_, _ = g.AddNode(infra.NewModule("B", "dep1"), g.Root(), nil)
		assert.Equal(t, len(g.QueryModule("A")), 1)
	})
}
