dest = "internal/3rdparty/protos"
sources = ["./protos"]

repo "https://github.com/grpc/grpc.git" {
  prefix = "grpc/"
  root = "src/proto/"
}

repo "https://github.com/protocolbuffers/protobuf.git" {
  prefix = "google/protobuf/"
  root = "src/"
}