package infra

type OperationEdge struct {
	From *OperationNode
	To   *OperationNode
}

type OperationNode struct {
	Operation Operation
}

type OperationGraph struct {
	// queries
	fromEdges map[*OperationNode][]*OperationEdge
	toEdges   map[*OperationNode][]*OperationEdge

	// data
	nodes []*OperationNode
}

func NewOperationGraph() *OperationGraph {
	return &OperationGraph{
		fromEdges: map[*OperationNode][]*OperationEdge{},
		toEdges:   map[*OperationNode][]*OperationEdge{},
	}
}

func (g *OperationGraph) AddNode(operation Operation) *OperationNode {
	node := &OperationNode{Operation: operation}
	g.nodes = append(g.nodes, node)
	return node
}

func (g *OperationGraph) DelNode(node *OperationNode) {
	for _, e := range g.In(node) {
		g.DelEdge(e)
	}
	for _, e := range g.Out(node) {
		g.DelEdge(e)
	}

	g.nodes = filter(g.nodes, node)
}

func (g *OperationGraph) In(node *OperationNode) []*OperationEdge {
	return g.toEdges[node]
}

func (g *OperationGraph) Out(node *OperationNode) []*OperationEdge {
	return g.fromEdges[node]
}

func (g *OperationGraph) DelEdge(edge *OperationEdge) {
	g.fromEdges[edge.From] = filter(g.fromEdges[edge.From], edge)
	g.toEdges[edge.To] = filter(g.toEdges[edge.To], edge)
	// Operation graph can have orphans
}

func (g *OperationGraph) FindNext() []*OperationNode {
	var result []*OperationNode
	for _, n := range g.nodes {
		if len(g.Out(n)) == 0 {
			result = append(result, n)
		}
	}
	return result
}
