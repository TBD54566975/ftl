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
key = apiKey.Get(ctx)
```

### Transforming secrets/configuration

Often, raw secret/configuration values aren't directly useful. For example, raw credentials might be used to create an API client. For those situations `ftl.Map()` can be used to transform a configuration or secret value into another type:

```go
var client = ftl.Map(ftl.Secret[Credentials]("credentials"),
                     func(ctx context.Context, creds Credentials) (*api.Client, error) {
    return api.NewClient(creds)
})
```
