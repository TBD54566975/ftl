Snippet for defining a value enum.

```go
//ftl:enum
type Animal interface { animal() }

type Cat struct {}
func (Cat) animal() {}
```
