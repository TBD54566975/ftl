Snippet for declaring a value enum.

```go
//ftl:enum
type Animal interface { animal() }

type Cat struct {}
func (Cat) animal() {}
```

See https://tbd54566975.github.io/ftl/docs/reference/types/
---
//ftl:enum
type ${1:Type} interface { ${2:interface}() }

type ${3:Value} struct {}
func (${3:Value}) ${2:interface}() {}
