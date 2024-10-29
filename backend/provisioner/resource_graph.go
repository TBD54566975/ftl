package provisioner

import (
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
)

// ResourceGraph is an in-memory graph of resources and their dependencies
type ResourceGraph struct {
	nodes []*provisioner.Resource
	edges []*ResourceEdge
}

type ResourceEdge struct {
	from string
	to   string
}

// AddNode to the graph
func (g *ResourceGraph) AddNode(n *provisioner.Resource) *provisioner.Resource {
	g.nodes = append(g.nodes, n)
	return n
}

// AddEdge between two nodes to the graph
func (g *ResourceGraph) AddEdge(from, to *provisioner.Resource) *ResourceEdge {
	edge := &ResourceEdge{from: from.ResourceId, to: to.ResourceId}
	g.edges = append(g.edges, edge)
	return edge
}

// Resources returns all nodes in the graph
func (g *ResourceGraph) Resources() []*provisioner.Resource {
	if g == nil {
		return nil
	}
	return g.nodes
}

// Roots returns all nodes that have no incoming edges
func (g *ResourceGraph) Roots() []*provisioner.Resource {
	var roots []*provisioner.Resource
	for _, node := range g.nodes {
		if len(g.In(node.ResourceId)) == 0 {
			roots = append(roots, node)
		}
	}
	return roots
}

// In edges of a node
func (g *ResourceGraph) In(id string) []*ResourceEdge {
	var upstream []*ResourceEdge
	for _, edge := range g.edges {
		if edge.to == id {
			upstream = append(upstream, edge)
		}
	}
	return upstream
}

// Out edges of a node
func (g *ResourceGraph) Out(id string) []*ResourceEdge {
	var downstream []*ResourceEdge
	for _, edge := range g.edges {
		if edge.from == id {
			downstream = append(downstream, edge)
		}
	}
	return downstream
}

// Node returns a node by id
func (g *ResourceGraph) Node(id string) *provisioner.Resource {
	for _, node := range g.nodes {
		if node.ResourceId == id {
			return node
		}
	}
	return nil
}

// Dependencies returns all downstream dependencies of a node
func (g *ResourceGraph) Dependencies(id string) []*provisioner.Resource {
	var deps []*provisioner.Resource
	for _, edge := range g.Out(id) {
		deps = append(deps, g.Node(edge.to))
	}
	return deps
}

// WithDirectDependencies returns a subgraph of given nodes with their direct dependencies
func (g *ResourceGraph) WithDirectDependencies(roots []*provisioner.Resource) *ResourceGraph {
	deps := map[string]bool{}
	edges := []*ResourceEdge{}
	for _, node := range roots {
		for _, e := range g.Out(node.ResourceId) {
			edges = append(edges, e)
			deps[e.to] = true
		}
	}
	nodes := []*provisioner.Resource{}
	for nd := range deps {
		nodes = append(nodes, g.Node(nd))
	}
	nodes = append(nodes, roots...)

	return &ResourceGraph{
		nodes: nodes,
		edges: edges,
	}
}

// ByIDs returns a slice of the resources with the given ids
func (g *ResourceGraph) ByIDs(ids map[string]bool) []*provisioner.Resource {
	var result []*provisioner.Resource
	for _, node := range g.nodes {
		if ids[node.ResourceId] {
			result = append(result, node)
		}
	}
	return result
}

// Update the state of existing resources
func (g *ResourceGraph) Update(resources []*provisioner.Resource) {
	byID := map[string]*provisioner.Resource{}
	for _, res := range resources {
		byID[res.ResourceId] = res
	}
	for i, node := range g.nodes {
		if n, ok := byID[node.ResourceId]; ok {
			g.nodes[i] = n
		}
	}
}
