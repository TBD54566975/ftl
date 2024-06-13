Snippet for defining a type enum (sum types).

```go
//ftl:enum
type MyEnum string

const (
	Value1 MyEnum = "Value1"
	Value2 MyEnum = "Value2"
)
```
---
//ftl:enum
type ${1:Enum} string

const (
	${2:Value1} ${1:Enum} = "${2:Value1}"
	${3:Value2} ${1:Enum} = "${3:Value2}"
)
