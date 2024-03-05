# Log File Navigator (Optional)

`lnav` is a [log file navigator](https://lnav.org) that is capable of reading
and analyzing log files in real time. It is a terminal application and is the a
robust tool for navigating and analyzing log files.

It's similar to `multitail` but more powerful more discoverable UX.

Some notable features:
 - Creates a unified view across multiple formats and types (e.g. JSON, syslog + `ftl` and `temporal`).
 - Persistent log sessions (e.g. filters, queries, and views).
 - Alternative querying via SQL (e.g. `SELECT * FROM log WHERE level = 'error'`).
 - Histogram view (e.g. log levels, time intervals) for a quick overview.
 - Unique coloring based on value types (e.g. multiple payloads will be colored differently).

## 1. Install `lnav` definition

```bash
.lnav/install.sh
```

## 2. Getting logs from `ftl`

```bash
ftl dev . \
	--log-level=debug \
	--log-json 2>&1 |
	tee -a "logs/ftl-$(date +'%F').log.json"
```

```bash
ftl serve \
	--allow-origins "*" \
	--log-level=debug \
	--log-json 2>&1 |
	tee -a "logs/ftl-$(date +'%F').log.json"
```

## 3. Analyzing logs

```bash
lnav -r logs/*
```
