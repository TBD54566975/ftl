Declare a value enum.

A value enum is an enumerated set of string or integer values.

```go
//ftl:enum
type Colour string

const (
  Red   Colour = "red"
  Green Colour = "green"
  Blue  Colour = "blue"
)
```

See https://tbd54566975.github.io/ftl/docs/reference/types/
---

//ftl:enum
type ${1:Enum} string

const (
	${2:Value1} ${1:Enum} = "${2:Value1}"
	${3:Value2} ${1:Enum} = "${3:Value2}"
)
