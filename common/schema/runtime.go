package schema

// RuntimeEvent is an event modifying a runtime part of the schema.
//
//sumtype:decl
type RuntimeEvent interface {
	runtimeEvent()
}
