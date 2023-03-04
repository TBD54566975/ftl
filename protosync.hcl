dest = "internal/3rdparty/protos"
sources = ["./internal/protos"]

repo "https://github.com/grpc/grpc.git" {
  prefix = "grpc/"
  root = "src/proto/"
}
