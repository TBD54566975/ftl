package infra

type Constraint struct {
	Key   string
	Value string
}

type ResourceNode struct {
	Resource Resource
}

type ResourceEdge struct {
	From        *ResourceNode
	To          *ResourceNode
	Constraints []Constraint
}

type ResourceGraph struct {
	// queries
	nodeByModule map[string][]*ResourceNode
	fromEdges    map[*ResourceNode][]*ResourceEdge
	toEdges      map[*ResourceNode][]*ResourceEdge

	// data
	root  *ResourceNode
	nodes []*ResourceNode
}

func NewResourceGraph() *ResourceGraph {
	ftl := &ResourceNode{Resource: &FTL{}}
	return &ResourceGraph{
		nodeByModule: map[string][]*ResourceNode{},
		fromEdges:    map[*ResourceNode][]*ResourceEdge{},
		toEdges:      map[*ResourceNode][]*ResourceEdge{},

		nodes: []*ResourceNode{ftl},
		root:  ftl,
	}
}

func (g *ResourceGraph) AddEdge(from, to *ResourceNode, constraints []Constraint) *ResourceEdge {
	edge := &ResourceEdge{
		From:        from,
		To:          to,
		Constraints: constraints,
	}

	g.fromEdges[from] = append(g.fromEdges[from], edge)
	g.toEdges[to] = append(g.toEdges[to], edge)
	return edge
}

func (g *ResourceGraph) DelEdge(edge *ResourceEdge) {
	g.fromEdges[edge.From] = filter(g.fromEdges[edge.From], edge)
	g.toEdges[edge.To] = filter(g.toEdges[edge.To], edge)

	if len(g.In(edge.To)) == 0 { // do not leave orphans behind
		g.delNode(edge.To)
	}
}

func (g *ResourceGraph) delNode(node *ResourceNode) {
	g.nodes = filter(g.nodes, node)
	g.nodeByModule[node.Resource.Id().Module] = filter(g.nodeByModule[node.Resource.Id().Module], node)
}

func (g *ResourceGraph) AddNode(resource Resource, parent *ResourceNode, constraints []Constraint) (*ResourceNode, *ResourceEdge) {
	node := &ResourceNode{Resource: resource}
	g.nodes = append(g.nodes, node)
	g.nodeByModule[node.Resource.Id().Module] = append(g.nodeByModule[node.Resource.Id().Module], node)
	edge := g.AddEdge(parent, node, constraints)
	return node, edge
}

func (g *ResourceGraph) Root() *ResourceNode {
	return g.root
}

func (g *ResourceGraph) In(node *ResourceNode) []*ResourceEdge {
	return g.toEdges[node]
}

func (g *ResourceGraph) Out(node *ResourceNode) []*ResourceEdge {
	return g.fromEdges[node]
}

func (g *ResourceGraph) QueryModule(name string) []*ResourceNode {
	return g.nodeByModule[name]
}

func (g *ResourceGraph) ById(id ResourceID) *ResourceNode {
	for _, v := range g.nodes {
		if v.Resource.Id() == id {
			return v
		}
	}
	return nil
}

func filter[T comparable](from []T, value T) []T {
	var res []T
	for _, v := range from {
		if v != value {
			res = append(res, v)
		}
	}
	return res
}
