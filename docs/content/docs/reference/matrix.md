+++
title = "Feature Matrix"
description = "Matrix showing which FTL features are supported by each language"
date = 2021-05-01T08:20:00+00:00
updated = 2021-05-01T08:20:00+00:00
draft = false
weight = 120
sort_by = "weight"
template = "docs/page.html"

[extra]
toc = true
top = false
+++

| System        | Feature         | Go  | JVM | Rust |
| :------------ | :-------------- | :-- | :-- | :--- |
| **Types**     | Basic Types     | ✔️  | ✔️  | ️ ✔️ |
|               | Optional Type   | ✔️  | ✔️  |      |
|               | Unit Type       | ✔️  | ✔️  |      |
|               | Empty Type      | ✔️  | ✔️  |      |
|               | Generic Types   | ✔️  | ️   |      |
|               | Type Aliases    | ✔️  | ️   |      |
|               | Value Enums     | ✔️  | ️   |      |
|               | Type Enums      | ✔️  | ️   |      |
|               | Visibility      | ✔️  | ✔️  |      |
| **Verbs**     | Verb            | ✔️  | ✔️  | ️✔️  |
|               | Sink            | ✔️  | ✔️  |      |
|               | Source          | ✔️  | ✔️  |      |
|               | Empty           | ✔️  | ✔️  |      |
|               | Visibility      | ✔️  | ✔️  |      |
| **Core**      | FSM             | ✔️  | ️   |      |
|               | Leases          | ✔️  | ✔️  |      |
|               | Cron            | ✔️  |     |      |
|               | Config          | ✔️  | ✔️  |      |
|               | Secrets         | ✔️  |     |      |
|               | HTTP Ingress    | ✔️  | ✔️  |      |
| **Resources** | PostgreSQL      | ✔️  | ️   |      |
|               | MySQL           |     |     |      |
|               | Kafka           |     |     |      |
| **PubSub**    | Declaring Topic | ✔️  | ✔️  |      |
|               | Subscribing     | ✔️  | ✔️  |      |
|               | Publishing      | ✔️  | ✔️  |      |
