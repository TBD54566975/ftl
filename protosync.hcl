dest = "internal/3rdparty/protos"
sources = ["./protos", "opentelemetry/proto/collector/metrics/v1/metrics_service.proto"]

repo "https://github.com/grpc/grpc.git" {
  prefix = "grpc/"
  root = "src/proto/"
}

repo "https://github.com/protocolbuffers/protobuf.git" {
  prefix = "google/protobuf/"
  root = "src/"
}

repo "https://github.com/open-telemetry/opentelemetry-proto.git" {
    prefix = "opentelemetry/proto/"
    root = ""
}