module github.com/TBD54566975/ftl

go 1.22.0

require (
	connectrpc.com/connect v1.15.0
	connectrpc.com/grpcreflect v1.2.0
	connectrpc.com/otelconnect v0.7.0
	github.com/BurntSushi/toml v1.3.2
	github.com/TBD54566975/scaffolder v0.8.0
	github.com/TBD54566975/scaffolder/extensions/javascript v0.8.0
	github.com/alecthomas/assert/v2 v2.6.0
	github.com/alecthomas/atomic v0.1.0-alpha2
	github.com/alecthomas/concurrency v0.0.2
	github.com/alecthomas/kong v0.8.1
	github.com/alecthomas/kong-toml v0.1.0
	github.com/alecthomas/participle/v2 v2.1.1
	github.com/alecthomas/types v0.13.0
	github.com/amacneil/dbmate/v2 v2.12.0
	github.com/beevik/etree v1.3.0
	github.com/bmatcuk/doublestar/v4 v4.6.1
	github.com/deckarep/golang-set/v2 v2.6.0
	github.com/go-logr/logr v1.4.1
	github.com/gofrs/flock v0.8.1
	github.com/golang/protobuf v1.5.3
	github.com/google/uuid v1.6.0
	github.com/jackc/pgerrcode v0.0.0-20220416144525-469b46aa5efa
	github.com/jackc/pgx/v5 v5.5.3
	github.com/jellydator/ttlcache/v3 v3.2.0
	github.com/jpillora/backoff v1.0.0
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/mattn/go-isatty v0.0.20
	github.com/oklog/ulid/v2 v2.1.0
	github.com/otiai10/copy v1.14.0
	github.com/radovskyb/watcher v1.0.7
	github.com/rs/cors v1.10.1
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1
	github.com/swaggest/jsonschema-go v0.3.66
	github.com/titanous/json5 v1.0.0
	github.com/tmc/langchaingo v0.1.5
	github.com/zalando/go-keyring v0.2.3
	go.opentelemetry.io/otel v1.24.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.24.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.24.0
	go.opentelemetry.io/otel/metric v1.24.0
	go.opentelemetry.io/otel/sdk v1.24.0
	go.opentelemetry.io/otel/sdk/metric v1.24.0
	go.opentelemetry.io/otel/trace v1.24.0
	go.uber.org/automaxprocs v1.5.3
	golang.org/x/exp v0.0.0-20240222234643-814bf88cf225
	golang.org/x/mod v0.15.0
	golang.org/x/net v0.21.0
	golang.org/x/sync v0.6.0
	golang.org/x/term v0.17.0
	golang.org/x/tools v0.18.0
	google.golang.org/protobuf v1.32.0
	modernc.org/sqlite v1.29.2
)

require (
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/pkoukk/tiktoken-go v0.1.2 // indirect
	go.opentelemetry.io/proto/otlp v1.1.0 // indirect
	modernc.org/gc/v3 v3.0.0-20240107210532-573471604cb6 // indirect
)

require (
	github.com/alecthomas/repr v0.4.0
	github.com/alessio/shellescape v1.4.2 // indirect
	github.com/benbjohnson/clock v1.3.5
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/danieljoos/wincred v1.2.0 // indirect
	github.com/dlclark/regexp2 v1.8.1 // indirect
	github.com/dop251/goja v0.0.0-20231027120936-b396bb4c349d // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/google/pprof v0.0.0-20230207041349-798e818bf904 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.0 // indirect
	github.com/hexops/gotextdiff v1.0.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/puzpuzpuz/xsync v1.5.2
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/serialx/hashring v0.0.0-20200727003509-22c0c7ab6b1b
	github.com/swaggest/refl v1.3.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.24.0 // indirect
	golang.design/x/reflect v0.0.0-20220504060917-02c43be63f3b
	golang.org/x/crypto v0.19.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240102182953-50ed04b92917 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240102182953-50ed04b92917 // indirect
	google.golang.org/grpc v1.61.1 // indirect
	modernc.org/libc v1.41.0 // indirect
	modernc.org/mathutil v1.6.0 // indirect
	modernc.org/memory v1.7.2 // indirect
	modernc.org/strutil v1.2.0 // indirect
	modernc.org/token v1.1.0 // indirect
)
