Snippet for declaring a FSM model.

```go
var payment = ftl.FSM(
  "payment",
  ftl.Start(Invoiced),
  ftl.Start(Paid),
  ftl.Transition(Invoiced, Paid),
  ftl.Transition(Invoiced, Defaulted),
)
```

See https://tbd54566975.github.io/ftl/docs/reference/fsm/
---
var ${1:FSM} = ftl.FSM(
  "${1:FSM}",
  ftl.Start(${2:verbState}),
  ftl.Transition(${2:fromVerbState}, ${3:toVerbState}),
)
```