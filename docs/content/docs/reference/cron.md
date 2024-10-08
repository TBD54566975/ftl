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

You can also use a shorthand syntax for the cron job, supporting seconds (`s`), minutes (`m`), hours (`h`), and specific days of the week (e.g. `Mon`).

### Examples

The following function will be called hourly:

{% code_selector() %}
<!-- go -->
```go
//ftl:cron 0 * * * *
func Hourly(ctx context.Context) error {
  // ...
}
```
<!-- kotlin -->
```kotlin
import xyz.block.ftl.Cron
@Cron("0 * * * *")
fun hourly() {
    
}
```
<!-- java -->
```java
import xyz.block.ftl.Cron;

class MyCron {
    @Cron("0 * * * *")
    void hourly() {
        
    }
}
```

{% end %}
Every 12 hours, starting at UTC midnight:

{% code_selector() %}
<!-- go -->
```go
//ftl:cron 12h
func TwiceADay(ctx context.Context) error {
  // ...
}
```
<!-- kotlin -->
```kotlin
import xyz.block.ftl.Cron
@Cron("12h")
fun twiceADay() {
    
}
```
<!-- java -->
```java
import xyz.block.ftl.Cron;

class MyCron {
    @Cron("12h")
    void twiceADay() {
        
    }
}
```
{% end %}

Every Monday at UTC midnight:

{% code_selector() %}
<!-- go -->
```go
//ftl:cron Mon
func Mondays(ctx context.Context) error {
  // ...
}
```
<!-- kotlin -->
```kotlin
import xyz.block.ftl.Cron
@Cron("Mon")
fun mondays() {
    
}
```
<!-- java -->
```java
import xyz.block.ftl.Cron;

class MyCron {
    @Cron("Mon")
    void mondays() {
        
    }
}
```
{% end %}