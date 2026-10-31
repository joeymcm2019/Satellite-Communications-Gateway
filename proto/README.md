# Shared protocol

`satellite/v1/satellite.proto` is the source of truth for the planned internal API. Fields use stable numbers and an unspecified enum value; timestamps use protobuf's standard Timestamp. Values such as battery percentages will need runtime validation when telemetry is implemented.

Bindings are deliberately not generated or imported yet. For the first gRPC milestone, install `protoc`, the Go `protoc-gen-go` and `protoc-gen-go-grpc` plugins, and Python `grpcio-tools` in the simulator virtual environment. Select and pin mutually compatible versions when introducing these dependencies.

From the repository root, the eventual manual commands are:

```sh
mkdir -p services/gateway/gen services/simulator/src/satellite/v1
protoc -I proto \
  --go_out=services/gateway/gen --go_opt=paths=source_relative \
  --go-grpc_out=services/gateway/gen --go-grpc_opt=paths=source_relative \
  proto/satellite/v1/satellite.proto
services/simulator/.venv/bin/python -m grpc_tools.protoc -I proto \
  --python_out=services/simulator/src \
  --pyi_out=services/simulator/src \
  --grpc_python_out=services/simulator/src \
  proto/satellite/v1/satellite.proto
```

The Go module path is a local placeholder and requires no GitHub repository. Update it and `go_package` together if a real remote is added. Generated Python modules will use `satellite.v1`; add package initializers and ensure setuptools discovers that namespace when wiring the client. Runtime dependencies `google.golang.org/grpc`, `google.golang.org/protobuf`, `grpcio`, and `protobuf` belong in the respective services at that stage. These commands have not been executed as part of this scaffold. No code-generation pipeline or Bazel configuration is present.
