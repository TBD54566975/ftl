package infra

type Operation interface {
	Apply(graph *ResourceGraph) error
	UpdateGraph(graph *ResourceGraph) (*ResourceGraph, error)
}

type OperationEdge struct {
	Blocker *Operation
	Blocked *Operation
}

type OperationNode struct {
	Operation Operation

	From *OperationEdge
	To   *OperationEdge
}

type OperationGraph struct {
	nodes []*OperationNode
}
