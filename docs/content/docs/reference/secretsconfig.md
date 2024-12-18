+++
title = "Secrets/Config"
description = "Secrets and Configuration values"
date = 2021-05-01T08:20:00+00:00
updated = 2021-05-01T08:20:00+00:00
draft = false
weight = 70
sort_by = "weight"
template = "docs/page.html"

[extra]
toc = true
top = false
+++

### Configuration

Configuration values are named, typed values. They are managed by the `ftl config` command-line.

{% code_selector() %}

<!-- go -->


To declare a configuration value use the following syntax:

```go
type DefaultUser = ftl.Config[Username]
```

Note that the name of the configuration value as represented in the FTL schema is the lower camel case version of the type name.

Configuration values can be injected into FTL methods, such as `@Verb`, HTTP ingress, Cron etc. To inject a configuration value, use the following syntax:

```go
//ftl:verb
func Hello(ctx context.Context, req Request, defaultUser DefaultUser) error {
    username := defaultUser.Get(ctx)
    // ...
}
```

<!-- kotlin -->

Configuration values can be injected into FTL methods, such as `@Verb`, HTTP ingress, Cron etc. To inject a configuration value, use the following syntax:

```kotlin
@Export
@Verb
fun hello(helloRequest: HelloRequest, @Config("defaultUser") defaultUser: String): HelloResponse {
    return HelloResponse("Hello, $defaultUser")
}
```
<!-- java -->
Configuration values can be injected into FTL methods, such as `@Verb`, HTTP ingress, Cron etc. To inject a configuration value, use the following syntax:

```java
@Export
@Verb
HelloResponse hello(HelloRequest helloRequest, @Config("defaultUser") String defaultUser)  {
    return new HelloResponse("Hello, " + defaultUser);
}
```

{% end %}


### Secrets

Secrets are encrypted, named, typed values. They are managed by the `ftl secret` command-line.

{% code_selector() %}

<!-- go -->

Declare a secret with the following:

```go
type ApiKey = ftl.Secret[Credentials]
```

Like configuration values, the name of the secret as represented in the FTL schema is the lower camel case version of the type name.

Configuration values can be injected into FTL methods, such as `@Verb`, HTTP ingress, Cron etc. To inject a configuration value, use the following syntax:

```go
//ftl:verb
func CallApi(ctx context.Context, req Request, apiKey ApiKey) error {
    credentials := apiKey.Get(ctx)
    // ...
}
```

<!-- kotlin -->

Configuration values can be injected into FTL methods, such as `@Verb`, HTTP ingress, Cron etc. To inject a configuration value, use the following syntax:

```kotlin
@Export
@Verb
fun hello(helloRequest: HelloRequest, @Secret("apiKey") apiKey: String): HelloResponse {
    return HelloResponse("Hello, ${api.call(apiKey)}")
}
```
<!-- java -->
Configuration values can be injected into FTL methods, such as `@Verb`, HTTP ingress, Cron etc. To inject a configuration value, use the following syntax:

```java
@Export
@Verb
HelloResponse hello(HelloRequest helloRequest, @Secret("apiKey") String apiKey)  {
    return new HelloResponse("Hello, " + api.call(apiKey));
}
```

{% end %}
### Transforming secrets/configuration

Often, raw secret/configuration values aren't directly useful. For example, raw credentials might be used to create an API client. For those situations `ftl.Map()` can be used to transform a configuration or secret value into another type:

```go
var client = ftl.Map(ftl.Secret[Credentials]("credentials"),
                     func(ctx context.Context, creds Credentials) (*api.Client, error) {
    return api.NewClient(creds)
})
```

This is not currently supported in Kotlin or Java.
