Declare a cron job, autocompleting on a cron expression.

A cron job is an Empty verb that will be called on a schedule. 

```go
//ftl:cron 0 * * * *
func Hourly(ctx context.Context) error {}

//ftl:cron 6h
func EverySixHours(ctx context.Context) error {}
```

See https://tbd54566975.github.io/ftl/docs/reference/cron/
---

//ftl:cron ${2:Minutes} ${3:Hours} ${4:DayOfMonth} ${5:Month} ${6:DayOfWeek}
func ${1:Name}(ctx context.Context) error {
	${7:// TODO: Implement}
	return nil
}
