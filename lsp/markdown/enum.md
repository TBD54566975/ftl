## Type enums (sum types)

[Sum types](https://en.wikipedia.org/wiki/Tagged_union) are supported by FTL's type system, but aren't directly supported by Go. However they can be approximated with the use of [sealed interfaces](https://blog.chewxy.com/2018/03/18/golang-interfaces/). To declare a sum type in FTL use the comment directive `//ftl:enum`:

```go
//ftl:enum
type Animal interface { animal() }

type Cat struct {}
func (Cat) animal() {}

type Dog struct {}
func (Dog) animal() {}
```

## Value enums

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

[Reference](https://tbd54566975.github.io/ftl/docs/reference/types/)
