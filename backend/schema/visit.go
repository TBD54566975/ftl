package schema

// Visit all nodes in the schema.
func Visit(n Node, visit func(n Node, next func() error) error) error {
	return visit(n, func() error {
		for _, child := range n.schemaChildren() {
			if err := Visit(child, visit); err != nil {
				return err
			}
		}
		return nil
	})
}

// VisitExcludingMetadataChildren visits all nodes in the schema except the children of metadata nodes.
// This is used when generating external modules to avoid adding imports only referenced in the bodies of
// stubbed verbs.
func VisitExcludingMetadataChildren(n Node, visit func(n Node, next func() error) error) error {
	return visit(n, func() error {
		if _, ok := n.(Metadata); !ok {
			for _, child := range n.schemaChildren() {
				if err := VisitExcludingMetadataChildren(child, visit); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
