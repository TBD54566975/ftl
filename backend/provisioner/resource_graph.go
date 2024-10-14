package provisioner

import "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"

// ResourceGraph is an in-memory graph of resources and their dependencies
type ResourceGraph struct {
	nodes []*provisioner.Resource
	edges []*ResourceEdge
}

type ResourceEdge struct {
	from *provisioner.Resource
	to   *provisioner.Resource
}

// AddNode to the graph
func (g *ResourceGraph) AddNode(n *provisioner.Resource) *provisioner.Resource {
	g.nodes = append(g.nodes, n)
	return n
}

// AddEdge between two nodes to the graph
func (g *ResourceGraph) AddEdge(from, to *provisioner.Resource) *ResourceEdge {
	edge := &ResourceEdge{from: from, to: to}
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
		if len(g.In(node)) == 0 {
			roots = append(roots, node)
		}
	}
	return roots
}

// In edges of a node
func (g *ResourceGraph) In(from *provisioner.Resource) []*ResourceEdge {
	var upstream []*ResourceEdge
	for _, edge := range g.edges {
		if edge.to == from {
			upstream = append(upstream, edge)
		}
	}
	return upstream
}

// Out edges of a node
func (g *ResourceGraph) Out(to *provisioner.Resource) []*ResourceEdge {
	var downstream []*ResourceEdge
	for _, edge := range g.edges {
		if edge.from == to {
			downstream = append(downstream, edge)
		}
	}
	return downstream
}

// Dependencies returns all downstream dependencies of a node
func (g *ResourceGraph) Dependencies(n *provisioner.Resource) []*provisioner.Resource {
	var deps []*provisioner.Resource
	for _, edge := range g.Out(n) {
		deps = append(deps, edge.to)
	}
	return deps
}

// WithDirectDependencies returns a subgraph of given nodes with their direct dependencies
func (g *ResourceGraph) WithDirectDependencies(roots []*provisioner.Resource) *ResourceGraph {
	deps := map[*provisioner.Resource]bool{}
	edges := []*ResourceEdge{}
	for _, node := range roots {
		for _, e := range g.Out(node) {
			edges = append(edges, e)
			deps[e.to] = true
		}
	}
	nodes := []*provisioner.Resource{}
	for nd := range deps {
		nodes = append(nodes, nd)
	}
	nodes = append(nodes, roots...)

	return &ResourceGraph{
		nodes: nodes,
		edges: edges,
	}
}
