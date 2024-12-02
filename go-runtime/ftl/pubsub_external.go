package ftl

// ExternalPartitionMapper is a fake partition mapper that is used to indicate that a topic is external.
// External topics can not be directly published to.
type ExternalPartitionMapper[E any] struct{}

var _ TopicPartitionMap[struct{}] = ExternalPartitionMapper[struct{}]{}

func (ExternalPartitionMapper[E]) PartitionKey(_ E) string {
	panic("directly publishing to external topics is not allowed")
}
