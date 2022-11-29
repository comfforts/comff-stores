# Comfforts store service

## Developer's notes

### Setup

- `git init`
- `go mod init`
- `touch api/v1/stores.proto`
- `go get google.golang.org/protobuf`
- `protoc api/v1/*.proto --go_out=. --go_opt=paths=source_relative --proto_path=.`
- `go get google.golang.org/grpc`
- `go get google.golang.org/grpc/cmd/protoc-gen-go-grpc`
- `protoc api/v1/*.proto --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --proto_path=.`
- setup initial server stub `internal/server/`
- setup geo hash util `pkg/utils/geohash/`
- setup store service `pkg/services/store/`
- `go get go.uber.org/zap`
- `gitlab.com/xerra/common/vincenty`
- `github.com/stretchr/testify`