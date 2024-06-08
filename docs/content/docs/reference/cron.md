+++
title = "Cron"
description = "Cron Jobs"
date = 2021-05-01T08:20:00+00:00
updated = 2021-05-01T08:20:00+00:00
draft = false
weight = 60
sort_by = "weight"
template = "docs/page.html"

[extra]
toc = true
top = false
+++

A cron job is an Empty verb that will be called on a schedule. The syntax is described [here](https://pubs.opengroup.org/onlinepubs/9699919799.2018edition/utilities/crontab.html).

eg. The following function will be called hourly:

```go
//ftl:cron 0 * * * *
func Hourly(ctx context.Context) error {
  // ...
}
```
