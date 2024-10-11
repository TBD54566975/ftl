# Metrics

FTL collects and exports a variety of metrics to help you monitor and understand the behavior of your FTL deployment using OTEL. This allows cluster operators to consume metrics in their preferred monitoring system e.g. Prometheus, Grafana, Datadog etc.

**Note on Metric Types:**
- **Counters** typically have a `_total` suffix in Prometheus, which is omitted here for brevity.
- **Histograms** typically have `_bucket`, `_sum`, and `_count` variants, with `_bucket` including multiple entries for different bucket boundaries.
- **Gauges** represent a single numerical value that can go up and down.

_Note: this documentation is incomplete_

## FTL Metrics

### Metrics Table

| Metric Name                          | Type      | Description                                        |
| ------------------------------------ | --------- | -------------------------------------------------- |
| `ftl.async_call.acquired`            | Counter   | Number of times an async call was acquired         |
| `ftl.async_call.completed`           | Counter   | Number of completed async calls                    |
| `ftl.async_call.created`             | Counter   | Number of created async calls                      |
| `ftl.async_call.executed`            | Counter   | Number of executed async calls                     |
| `ftl.async_call.ms_to_complete`      | Histogram | Time taken to complete async calls in milliseconds |
| `ftl.async_call.queue_depth_ratio`   | Gauge     | Ratio of queued async calls                        |
| `ftl.call.ms_to_complete`            | Histogram | Time taken to complete calls in milliseconds       |
| `ftl.call.requests`                  | Counter   | Total number of call requests                      |
| `ftl.deployments.runner.active`      | Gauge     | Number of active deployment runners                |
| `ftl.runner.registration.heartbeats` | Counter   | Total number of runner registration heartbeats     |
| `ftl.timeline.inserted`              | Counter   | Total number of timeline insertions                |

### Attributes

#### ftl.async_call.catching
- `ftl.async_call.acquired`
- `ftl.async_call.completed`
- `ftl.async_call.created`
- `ftl.async_call.executed`
- `ftl.async_call.ms_to_complete`

#### ftl.async_call.origin
- `ftl.async_call.acquired`
- `ftl.async_call.completed`
- `ftl.async_call.created`
- `ftl.async_call.executed`
- `ftl.async_call.ms_to_complete`

#### ftl.async_call.time_since_scheduled_at_ms.bucket
- `ftl.async_call.acquired`
- `ftl.async_call.completed`
- `ftl.async_call.executed`

#### ftl.async_call.verb.ref
- `ftl.async_call.acquired`
- `ftl.async_call.completed`
- `ftl.async_call.created`
- `ftl.async_call.executed`
- `ftl.async_call.ms_to_complete`

#### ftl.module.name
- `ftl.async_call.acquired`
- `ftl.async_call.completed`
- `ftl.async_call.created`
- `ftl.async_call.executed`
- `ftl.async_call.ms_to_complete`
- `ftl.call.ms_to_complete`
- `ftl.call.requests`

#### ftl.outcome.status
- `ftl.async_call.acquired`
- `ftl.async_call.completed`
- `ftl.async_call.created`
- `ftl.async_call.executed`
- `ftl.async_call.ms_to_complete`
- `ftl.call.ms_to_complete`
- `ftl.call.requests`

#### ftl.async_call.remaining_attempts
- `ftl.async_call.created`

#### ftl.call.verb.ref
- `ftl.call.ms_to_complete`
- `ftl.call.requests`

#### ftl.call.run_time_ms.bucket
- `ftl.call.requests`

#### ftl.deployment.key
- `ftl.deployments.runner.active`
- `ftl.runner.registration.heartbeats`

## DB Metrics

DB Metrics are collected from the SQL database used by FTL. These metrics provide insights into the database connection pool, query latency etc.

### Metrics Table

| Metric Name                                    | Type      | Description                                                    |
| ---------------------------------------------- | --------- | -------------------------------------------------------------- |
| `db.sql.connection.closed_max_idle_time`       | Counter   | Total number of connections closed due to max idle time        |
| `db.sql.connection.closed_max_idle`            | Counter   | Total number of connections closed due to max idle connections |
| `db.sql.connection.closed_max_lifetime`        | Counter   | Total number of connections closed due to max lifetime         |
| `db.sql.connection.max_open`                   | Gauge     | Maximum number of open connections                             |
| `db.sql.connection.open`                       | Gauge     | Number of open connections                                     |
| `db.sql.connection.wait_duration_milliseconds` | Histogram | Duration of connection wait times in milliseconds              |
| `db.sql.connection.wait`                       | Counter   | Total number of connection waits                               |
| `db.sql.latency_milliseconds`                  | Histogram | SQL query latency in milliseconds                              |

### Attributes

#### db.system
- `db.sql.connection.closed_max_idle_time`
- `db.sql.connection.closed_max_idle`
- `db.sql.connection.closed_max_lifetime`
- `db.sql.connection.max_open`
- `db.sql.connection.open`
- `db.sql.connection.wait_duration_milliseconds`
- `db.sql.connection.wait`

#### status
- `db.sql.connection.open`
- `db.sql.latency_milliseconds`

#### method
- `db.sql.latency_milliseconds`

#### le (less than or equal to)
- `db.sql.latency_milliseconds_bucket`

Note: The `job` attribute with value "ftl-serve" is common to all metrics and has been omitted from the individual listings for brevity.