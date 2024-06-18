Declare a secret.

Secrets are encrypted, named, typed values. They are managed by the `ftl secret` command-line.

```go
var apiKey = ftl.Secret[string]("apiKey")
```

See https://tbd54566975.github.io/ftl/docs/reference/secretsconfig/
---

var ${1:secretVar} = ftl.Secret[${2:Type}]("${1:secretVar}")
