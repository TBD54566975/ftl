package schema

import "github.com/alecthomas/errors"

// Visit all nodes in the schema.
func Visit(n Node, visit func(n Node, next func() error) error) error {
	return visit(n, func() error {
		for _, child := range n.schemaChildren() {
			if err := Visit(child, visit); err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	})
}
