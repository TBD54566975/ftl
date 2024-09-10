package infra

type Constraint struct {
	Key   string
	Value string
}

type ResourceNode struct {
	Properties map[string]string
}

func (n *ResourceNode) Id() string {
	return "[" + n.Properties["kind"] + "," + n.Properties["module"] + "," + n.Properties["deployment"] + "]"
}

type ResourceEdge struct {
	From        *ResourceNode
	To          *ResourceNode
	Constraints []Constraint
}

type ResourceGraph struct {
	// queries
	fromEdges map[*ResourceNode][]*ResourceEdge
	toEdges   map[*ResourceNode][]*ResourceEdge

	// data
	root  *ResourceNode
	nodes []*ResourceNode
}

func NewResourceGraph() *ResourceGraph {
	ftl := &ResourceNode{Properties: map[string]string{"kind": "FTL"}}
	return &ResourceGraph{
		fromEdges: map[*ResourceNode][]*ResourceEdge{},
		toEdges:   map[*ResourceNode][]*ResourceEdge{},

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

func (g *ResourceGraph) Edge(from, to *ResourceNode) *ResourceEdge {
	for _, e := range g.fromEdges[from] {
		if e.To == to {
			return e
		}
	}
	return nil
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
}

func (g *ResourceGraph) DelNode(node *ResourceNode) {
	if len(g.Out(node)) > 0 {
		panic("trying to remove a node with children")
	}

	for _, e := range g.In(node) {
		g.DelEdge(e)
	}
	g.delNode(node)
}

func (g *ResourceGraph) AddNode(properties map[string]string, parent *ResourceNode, constraints []Constraint) (*ResourceNode, *ResourceEdge) {
	node := &ResourceNode{Properties: properties}
	g.nodes = append(g.nodes, node)
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

func (g *ResourceGraph) ById(id string) *ResourceNode {
	for _, v := range g.nodes {
		if v.Id() == id {
			return v
		}
	}
	return nil
}

func (g *ResourceGraph) Find(matcher Matcher) []*ResourceNode {
	var result []*ResourceNode
	for _, v := range g.nodes {
		if matcher(v, g) {
			result = append(result, v)
		}
	}
	return result
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

// Merge other graph o to g. This is a destructive operation on o
func (g *ResourceGraph) Merge(o *ResourceGraph) {
	var edges []*ResourceEdge
	for _, n := range o.nodes {
		existing := g.ById(n.Id())
		if existing == nil {
			g.nodes = append(g.nodes, n)
		} else {
			mergeProperties(existing, n)
		}
		edges = append(edges, o.fromEdges[n]...)
	}

	for _, e := range edges {
		from := g.ById(e.From.Id())
		to := g.ById(e.To.Id())
		if g.Edge(from, to) == nil {
			g.AddEdge(from, to, e.Constraints)
		} else {
			// TODO: Clever edge merge stuff
		}
	}
}

func mergeProperties(to, from *ResourceNode) {
	// TODO: Clever node merge stuff
	for k, v := range from.Properties {
		to.Properties[k] = v
	}
}
