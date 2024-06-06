# FTL Reference

## Import the runtime

Some aspects of FTL rely on a runtime which must be imported with:

```go
import "github.com/TBD54566975/ftl/go-runtime/ftl"
```

## Verbs

### Defining Verbs

To declare a Verb, write a normal Go function with the following signature,annotated with the Go [comment directive](https://tip.golang.org/doc/comment#syntax) `//ftl:verb`:

```go
//ftl:verb
func F(context.Context, In) (Out, error) { }
```

eg.

```go
type EchoRequest struct {}

type EchoResponse struct {}

//ftl:verb
func Echo(ctx context.Context, in EchoRequest) (EchoResponse, error) {
  // ...
}
```

By default verbs are only [visible](#Visibility) to other verbs in the same module.

### Calling Verbs

To call a verb use `ftl.Call()`. eg.

```go
out, err := ftl.Call(ctx, echo.Echo, echo.EchoRequest{})
```

## Types

FTL supports the following types: `Int` (64-bit), `Float` (64-bit), `String`, `Bytes` (a byte array), `Bool`, `Time`, `Any` (a dynamic type), `Unit` (similar to "void"), arrays, maps, data structures, and constant enumerations. Each FTL type is mapped to a corresponding language-specific type. For example in Go `Float` is represented as `float64`, `Time` is represented by `time.Time`, and so on. [^2]

Any Go type supported by FTL and referenced by an FTL declaration will be automatically exposed to an FTL type.

For example, the following verb declaration will result in `Request` and `Response` being automatically translated to FTL types.

```go
type Request struct {}
type Response struct {}

//ftl:verb
func Hello(ctx context.Context, in Request) (Response, error) {
  // ...
}
```

### Type enums (sum types)

[Sum types](https://en.wikipedia.org/wiki/Tagged_union) are supported by FTL's type system, but aren't directly supported by Go. However they can be approximated with the use of [sealed interfaces](https://blog.chewxy.com/2018/03/18/golang-interfaces/). To declare a sum type in FTL use the comment directive `//ftl:enum`:

```go
//ftl:enum
type Animal interface { animal() }

type Cat struct {}
func (Cat) animal() {}

type Dog struct {}
func (Dog) animal() {}
```

### Value enums

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

### Type aliases

A type alias is an alternate name for an existing type. It can be declared like so:

```go
//ftl:typealias
type Alias Target
```

eg.

```go
//ftl:typealias
type UserID string
```

## Visibility

By default all declarations in FTL are visible only to the module they're declared in. The implicit visibility of types is that of the first verb or other declaration that references it.

### Exporting declarations

Exporting a declaration makes it accessible to other modules. Some declarations that are entirely local to a module, such as secrets/config, cannot be exported.

Types that are transitively referenced by an exported declaration will be automatically exported unless they were already defined but unexported. In this case, an error will be raised and the type must be explicitly exported.

The following table describes the directives used to export the corresponding declaration:

| Symbol   | Export syntax        |
| -------- | -------------------- |
| Verb     | `//ftl:verb export`  |
| Data     | `//ftl:data export`  |
| Enum/Sum type | `//ftl:enum export`  |
| Typealias| `//ftl:typealias export` |
| Topic    | `//ftl:export` [^1]  |

eg.

```go
//ftl:verb export
func Verb(ctx context.Context, in In) (Out, error)

//ftl:typealias export
type UserID string
```

## HTTP ingress

Verbs annotated with `ftl:ingress` will be exposed via HTTP (`http` is the default ingress type). These endpoints will then be available on one of our default `ingress` ports (local development defaults to `http://localhost:8891`).

The following will be available at `http://localhost:8891/http/users/123/posts?postId=456`.

```go
type GetRequest struct {
	UserID string `json:"userId"`
	PostID string `json:"postId"`
}

type GetResponse struct {
	Message string `json:"msg"`
}

//ftl:ingress GET /http/users/{userId}/posts
func Get(ctx context.Context, req builtin.HttpRequest[GetRequest]) (builtin.HttpResponse[GetResponse, ErrorResponse], error) {
  // ...
}
```

> [!Important]
> The `req` and `resp` types of HTTP `ingress` [verbs](#Verb) must be `builtin.HttpRequest` and `builtin.HttpResponse` respectively. These types provide the necessary fields for HTTP `ingress` (`headers`, `statusCode`, etc.)
> 

Key points to note
* `path`, `query`, and `body` parameters are automatically mapped to the `req` and `resp` structures. In the example above, `{userId}` is extracted from the path parameter and `postId` is extracted from the query parameter.
* `ingress` verbs will be automatically exported by default.

## Cron jobs

A cron job is an Empty verb that will be called on a schedule. The syntax is described [here](https://pubs.opengroup.org/onlinepubs/9699919799.2018edition/utilities/crontab.html).

eg. The following function will be called hourly:

```go
//ftl:cron 0 * * * *
func Hourly(ctx context.Context) error {
  // ...
}
```

## Secrets/configuration

### Configuration

Configuration values are named, typed values. They are managed by the `ftl config` command-line.

To declare a configuration value use the following syntax:

```go
var defaultUser = ftl.Config[string]("default")
```

Then to retrieve a configuration value:

```go
username = defaultUser.Get(ctx)
```

### Secrets

Secrets are encrypted, named, typed values. They are managed by the `ftl secret` command-line.

Declare a secret with the following:

```go
var apiKey = ftl.Secret[string]("apiKey")
```

Then to retrieve a secret value:

```go
username = defaultUser.Get(ctx)
```

### Transforming secrets/configuration

Often, raw secret/configuration values aren't directly useful. For example, raw credentials might be used to create an API client. For those situations `ftl.Map()` can be used to transform a configuration or secret value into another type:

```go
var client = ftl.Map(ftl.Secret[Credentials]("credentials"),
                     func(ctx context.Context, creds Credentials) (*api.Client, error) {
    return api.NewClient(creds)
})
```

## PubSub

PubSub is a first-class concept in FTL, modelled on the concepts of topics (where events are sent), subscriptions (a cursor over the topic), and subscribers (where events are delivered to). Susbcribers are, as you would expect, Sinks. Each subscription is a cursor over the topic it is associated with. Each topic may have multiple subscriptions. Each subscription may have multiple subscribers, in which case events will be distributed among them.

First, declare a new topic:

```go
var invoicesTopic = ftl.Topic[Invoice]("invoices")
```

Then declare each subscription on the topic:

```go
var _ = ftl.Subscription(invoicesTopic, "emailInvoices")
```

And finally define a Sink to consume from the subscription:

```go
//ftl:subscribe emailInvoices
func SendInvoiceEmail(ctx context.Context, in Invoice) error {
  // ...
}
```

## FSM

FTL has first-class support for distributed [finite-state machines](https://en.wikipedia.org/wiki/Finite-state_machine). Each state in the state machine is a Sink, with events being values of the type of each sinks input. The FSM is declared once, with each executing instance of the FSM identified by a unique key when sending an event to it.

Here's an example of an FSM that models a simple payment flow:

```go
var payment = ftl.FSM(
  "payment",
  ftl.Start(Invoiced),
  ftl.Start(Paid),
  ftl.Transition(Invoiced, Paid),
  ftl.Transition(Invoiced, Defaulted),
)

//ftl:verb
func SendDefaulted(ctx context.Context, in DefaultedInvoice) error {
  return payment.Send(ctx, in.InvoiceID, in.Timeout)
}

//ftl:verb
func Invoiced(ctx context.Context, in Invoice) error { 
  if timedOut {
    return ftl.CallAsync(ctx, SendDefaulted, Timeout{...})
  }
}

//ftl:verb
func Paid(ctx context.Context, in Receipt) error { /* ... */ }

//ftl:verb
func Defaulted(ctx context.Context, in Timeout) error { /* ... */ }
```

Then to send events to the FSM:

```go
err := payment.Send(ctx, invoiceID, Invoice{Amount: 110})
```

Sending an event to an FSM is asynchronous. From the time an event is sent until the state function completes execution, the FSM is transitioning. It is invalid to send an event to an FSM that is transitioning.

## Retries

Any verb called asynchronously (specifically, PubSub subscribers and FSM states), may optionally specify a basic exponential backoff retry policy via a Go comment directive. The directive has the following syntax:

```go
//ftl:retry [<attempts>] <min-backoff> [<max-backoff>]
```

`attempts` and `max-backoff` default to unlimited if not specified.

For example, the following function will retry up to 10 times, with a delay of 5s, 10s, 20s, 40s, 60s, 60s, etc.

```go
//ftl:retry 10 5s 1m
func Invoiced(ctx context.Context, in Invoice) error {
  // ...
}
```

*[Verb]: func(context.Context, In) (Out, error)
*[Verbs]: func(context.Context, In) (Out, error)
*[Sink]: func(context.Context, In) error
*[Sinks]: func(context.Context, In) error
*[Source]: func(context.Context) (Out, error)
*[Sources]: func(context.Context) (Out, error)
*[Empty]: func(context.Context) error

[^1]: Annotation of topics is usually unnecessary.
[^2]: Note that until [type widening](https://github.com/TBD54566975/ftl/issues/1296) is implemented, external types are not supported.
