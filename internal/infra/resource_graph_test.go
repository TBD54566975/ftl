package infra_test

import (
	"testing"

	"github.com/TBD54566975/ftl/internal/infra"
	"gotest.tools/v3/assert"
)

func TestRGEdges(t *testing.T) {
	t.Run("returns incoming edges", func(t *testing.T) {
		g := infra.NewResourceGraph()
		deployment, _ := g.AddNode(map[string]string{"kind": "deployment"}, g.Root(), nil)
		assert.Equal(t, len(g.In(deployment)), 1)
	})
	t.Run("returns outgoing edges", func(t *testing.T) {
		g := infra.NewResourceGraph()
		_, _ = g.AddNode(map[string]string{"kind": "deployment"}, g.Root(), nil)
		_, _ = g.AddNode(map[string]string{"kind": "deployment"}, g.Root(), nil)
		assert.Equal(t, len(g.Out(g.Root())), 2)
	})
	t.Run("deleting last incoming edge removes the node", func(t *testing.T) {
		g := infra.NewResourceGraph()
		node, e1 := g.AddNode(map[string]string{"kind": "module", "module": "A"}, g.Root(), nil)

		e2 := g.AddEdge(g.Root(), node, nil)
		assert.Assert(t, g.ById(node.Id()) != nil)

		g.DelEdge(e1)
		assert.Assert(t, g.ById(node.Id()) != nil)

		g.DelEdge(e2)
		assert.Assert(t, g.ById(node.Id()) == nil)
	})
}

func TestRGQueries(t *testing.T) {
	t.Run("queries by module", func(t *testing.T) {
		g := infra.NewResourceGraph()
		_, _ = g.AddNode(map[string]string{"kind": "module", "module": "A"}, g.Root(), nil)
		_, _ = g.AddNode(map[string]string{"kind": "module", "module": "B"}, g.Root(), nil)
		assert.Equal(t, len(g.Find(infra.ByModule("A"))), 1)
	})
}

func TestRGDeployment(t *testing.T) {
	t.Run("deploys a single module", func(t *testing.T) {
		deployment := "depid"
		g := infra.NewResourceGraph()
		moduleA, _ := g.AddNode(map[string]string{"kind": "module", "module": "A", "deployment": deployment}, g.Root(), nil)
		moduleA.Properties["state"] = "planning"

		assert.DeepEqual(t, infra.Rules.ApplyAll(g), []string{
			"run provision-module [module,A,depid]",
		})
	})
	t.Run("deploys more complex plan", func(t *testing.T) {
		deployment := "depid"
		g := infra.NewResourceGraph()

		deployGraph := infra.NewResourceGraph()
		moduleB, _ := deployGraph.AddNode(map[string]string{"kind": "module", "module": "B", "deployment": deployment, "state": "planning"}, g.Root(), nil)
		moduleA, _ := deployGraph.AddNode(map[string]string{"kind": "module", "module": "A", "deployment": deployment, "state": "planning"}, g.Root(), nil)
		deployGraph.AddEdge(moduleB, moduleA, nil)
		_, _ = deployGraph.AddNode(map[string]string{"kind": "database", "module": "A", "deployment": deployment, "state": "planning"}, moduleA, nil)

		g.Merge(deployGraph)

		assert.DeepEqual(t, infra.Rules.ApplyAll(g), []string{
			"run provision-database [database,A,depid]",
			"run provision-module [module,A,depid]",
			"run provision-module [module,B,depid]",
		})
	})
	t.Run("supports concurrent deploys", func(t *testing.T) {
		deployment := "depid"
		g := infra.NewResourceGraph()

		deployGraph1 := infra.NewResourceGraph()
		moduleB, _ := deployGraph1.AddNode(map[string]string{"kind": "module", "module": "B", "deployment": deployment, "state": "planning"}, g.Root(), nil)
		moduleA, _ := deployGraph1.AddNode(map[string]string{"kind": "module", "module": "A", "deployment": deployment, "state": "planning"}, g.Root(), nil)
		deployGraph1.AddEdge(moduleB, moduleA, nil)
		_, _ = deployGraph1.AddNode(map[string]string{"kind": "database", "module": "A", "deployment": deployment, "state": "planning"}, moduleA, nil)

		// deploy a change to moduleB requiring a DB
		deployGraph2 := infra.NewResourceGraph()
		moduleB2, _ := deployGraph2.AddNode(map[string]string{"kind": "module", "module": "B", "deployment": deployment, "state": "planning"}, g.Root(), nil)
		_, _ = deployGraph2.AddNode(map[string]string{"kind": "database", "module": "B", "deployment": deployment, "state": "planning"}, moduleB2, nil)

		g.Merge(deployGraph1)
		g.Merge(deployGraph2)

		assert.DeepEqual(t, infra.Rules.ApplyAll(g), []string{
			"run provision-database [database,A,depid]",
			"run provision-module [module,A,depid]",
			"run provision-database [database,B,depid]",
			"run provision-module [module,B,depid]",
		})
	})
	t.Run("deploys module removal", func(t *testing.T) {
		deployment := "depid"
		g := infra.NewResourceGraph()

		dg := infra.NewResourceGraph()
		b, _ := dg.AddNode(map[string]string{"kind": "module", "module": "B", "deployment": deployment, "state": "planning"}, g.Root(), nil)
		_, _ = dg.AddNode(map[string]string{"kind": "database", "module": "B", "deployment": deployment, "state": "planning"}, b, nil)

		g.Merge(dg)
		infra.Rules.ApplyAll(g)

		dg = infra.NewResourceGraph()
		_, _ = dg.AddNode(map[string]string{"kind": "module", "module": "B", "deployment": deployment, "state": "outdated"}, g.Root(), nil)
		g.Merge(dg)

		assert.DeepEqual(t, infra.Rules.ApplyAll(g), []string{
			"run delete-module [module,B,depid]",
			"outdate [database,B,depid]",
			"run delete-database [database,B,depid]",
			"delete [database,B,depid]",
			"delete [module,B,depid]",
		})
	})
	t.Run("does not remove a module with dependencies", func(t *testing.T) {
		deployment := "depid"
		g := infra.NewResourceGraph()

		dg := infra.NewResourceGraph()
		moduleB, _ := dg.AddNode(map[string]string{"kind": "module", "module": "B", "deployment": deployment, "state": "planning"}, g.Root(), nil)
		moduleA, _ := dg.AddNode(map[string]string{"kind": "module", "module": "A", "deployment": deployment, "state": "planning"}, g.Root(), nil)
		dg.AddEdge(moduleB, moduleA, nil)

		g.Merge(dg)
		infra.Rules.ApplyAll(g)

		dg = infra.NewResourceGraph()
		_, _ = dg.AddNode(map[string]string{"kind": "module", "module": "A", "deployment": deployment, "state": "outdated"}, g.Root(), nil)
		g.Merge(dg)

		assert.Equal(t, len(infra.Rules.ApplyAll(g)), 0)
	})
}
