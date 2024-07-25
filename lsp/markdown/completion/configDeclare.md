Declare a config variable.

Configuration values are named, typed values. They are managed by the `ftl config` command-line.

```go
var defaultUser = ftl.Config[string]("defaultUser")
```

See https://tbd54566975.github.io/ftl/docs/reference/secretsconfig/
---

var ${1:configVar} = ftl.Config[${2:Type}]("${1:configVar}")
