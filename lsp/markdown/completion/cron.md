Snippet for declaring a cron job.

```go
//ftl:cron 0 * * * *
func Hourly(ctx context.Context) error {}

//ftl:cron 6h
func EverySixHours(ctx context.Context) error {}
```

See https://tbd54566975.github.io/ftl/docs/reference/cron/
---

//ftl:cron ${1:Schedule}
func ${2:Name}(ctx context.Context) error {
	${3:// TODO: Implement}
	return nil
}
```