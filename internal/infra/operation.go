package infra

// Operation can have a side, effect, change the reource graph, or both.
//
// Side effect is executed first. When planning changes, only resource graph updates are executed
type Operation interface {
	RunSideEffects() error
	UpdateGraph(graph *ResourceGraph) error
}
