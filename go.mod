module github.com/TBD54566975/ftl

go 1.23.0

require (
	connectrpc.com/connect v1.16.2
	connectrpc.com/grpcreflect v1.2.0
	connectrpc.com/otelconnect v0.7.1
	github.com/BurntSushi/toml v1.4.0
	github.com/TBD54566975/golang-tools v0.2.1
	github.com/TBD54566975/scaffolder v1.1.0
	github.com/XSAM/otelsql v0.33.0
	github.com/alecthomas/assert/v2 v2.10.0
	github.com/alecthomas/atomic v0.1.0-alpha2
	github.com/alecthomas/chroma/v2 v2.14.0
	github.com/alecthomas/concurrency v0.0.2
	github.com/alecthomas/kong v0.9.0
	github.com/alecthomas/kong-toml v0.2.0
	github.com/alecthomas/participle/v2 v2.1.1
	github.com/alecthomas/types v0.16.0
	github.com/amacneil/dbmate/v2 v2.20.0
	github.com/aws/aws-sdk-go v1.55.5
	github.com/aws/aws-sdk-go-v2 v1.30.5
	github.com/aws/aws-sdk-go-v2/config v1.27.33
	github.com/aws/aws-sdk-go-v2/credentials v1.17.32
	github.com/aws/aws-sdk-go-v2/service/kms v1.35.7
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.32.8
	github.com/aws/smithy-go v1.20.4
	github.com/beevik/etree v1.4.1
	github.com/bmatcuk/doublestar/v4 v4.6.1
	github.com/deckarep/golang-set/v2 v2.6.0
	github.com/docker/docker v27.2.1+incompatible
	github.com/docker/go-connections v0.5.0
	github.com/go-logr/logr v1.4.2
	github.com/google/uuid v1.6.0
	github.com/hashicorp/cronexpr v1.1.2
	github.com/jackc/pgerrcode v0.0.0-20240316143900-6e2875d9b438
	github.com/jackc/pgx/v5 v5.7.0
	github.com/jellydator/ttlcache/v3 v3.3.0
	github.com/jpillora/backoff v1.0.0
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/mattn/go-isatty v0.0.20
	github.com/multiformats/go-base36 v0.2.0
	github.com/otiai10/copy v1.14.0
	github.com/radovskyb/watcher v1.0.7
	github.com/rs/cors v1.11.1
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1
	github.com/sqlc-dev/pqtype v0.3.0
	github.com/swaggest/jsonschema-go v0.3.72
	github.com/tink-crypto/tink-go-awskms v0.0.0-20230616072154-ba4f9f22c3e9
	github.com/tink-crypto/tink-go/v2 v2.2.0
	github.com/titanous/json5 v1.0.0
	github.com/tliron/commonlog v0.2.18
	github.com/tliron/glsp v0.2.2
	github.com/tliron/kutil v0.3.26
	github.com/tmc/langchaingo v0.1.12
	github.com/zalando/go-keyring v0.2.5
	go.opentelemetry.io/otel v1.29.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.29.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.29.0
	go.opentelemetry.io/otel/metric v1.29.0
	go.opentelemetry.io/otel/sdk v1.29.0
	go.opentelemetry.io/otel/sdk/metric v1.29.0
	go.opentelemetry.io/otel/trace v1.29.0
	go.uber.org/automaxprocs v1.5.3
	golang.org/x/exp v0.0.0-20240808152545-0cdaa3abc0fa
	golang.org/x/mod v0.21.0
	golang.org/x/net v0.29.0
	golang.org/x/sync v0.8.0
	golang.org/x/term v0.24.0
	google.golang.org/protobuf v1.34.2
	gotest.tools/v3 v3.5.1
	k8s.io/api v0.31.0
	k8s.io/apimachinery v0.31.0
	k8s.io/client-go v0.31.0
	modernc.org/sqlite v1.33.0
)

require (
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.22.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.26.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.30.7 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/distribution/reference v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/gorilla/websocket v1.5.1 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/muesli/termenv v0.15.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/onsi/gomega v1.33.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/petermattis/goid v0.0.0-20240813172612-4fcff4a6cae7 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkoukk/tiktoken-go v0.1.6 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/sasha-s/go-deadlock v0.3.5 // indirect
	github.com/segmentio/ksuid v1.0.4 // indirect
	github.com/sourcegraph/jsonrpc2 v0.2.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.53.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.26.0 // indirect
	go.opentelemetry.io/proto/otlp v1.3.1 // indirect
	golang.org/x/oauth2 v0.22.0 // indirect
	golang.org/x/time v0.6.0 // indirect
	golang.org/x/tools v0.24.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20240228011516-70dd3763d340 // indirect
	k8s.io/utils v0.0.0-20240711033017-18e509b52bc8 // indirect
	modernc.org/gc/v3 v3.0.0-20240107210532-573471604cb6 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

require (
	github.com/alecthomas/repr v0.4.0
	github.com/alessio/shellescape v1.4.2 // indirect
	github.com/benbjohnson/clock v1.3.5
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/danieljoos/wincred v1.2.0 // indirect
	github.com/dlclark/regexp2 v1.11.4 // indirect
	github.com/dop251/goja v0.0.0-20240816181238-8130cadc5774 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/google/pprof v0.0.0-20240525223248-4bfdf5a9a2af // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.22.0 // indirect
	github.com/hexops/gotextdiff v1.0.3
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/lib/pq v1.10.9
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/puzpuzpuz/xsync/v3 v3.4.0
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/swaggest/refl v1.3.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.29.0 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/sys v0.25.0
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240822170219-fc7c04adadcd // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240822170219-fc7c04adadcd // indirect
	google.golang.org/grpc v1.65.0 // indirect
	modernc.org/mathutil v1.6.0 // indirect
	modernc.org/memory v1.8.0 // indirect
	modernc.org/strutil v1.2.0 // indirect
	modernc.org/token v1.1.0 // indirect
)

retract (
	v1.1.5
	v1.1.4
	v1.1.3
	v1.1.2
	v1.1.1
	v1.1.0
)
