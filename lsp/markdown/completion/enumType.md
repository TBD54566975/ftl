Declare a type enum (sum types).

A type enum is a set of types that can be used as a single type, which `go` does not directly support.

```go
//ftl:enum
type Animal interface { animal() }

type Cat struct {}
func (Cat) animal() {}

type Dog struct {}
func (Dog) animal() {}
```

See https://tbd54566975.github.io/ftl/docs/reference/types/
---

//ftl:enum
type ${1:Type} interface { ${2:interface}() }

type ${3:Value} struct {}
func (${3:Value}) ${2:interface}() {}
