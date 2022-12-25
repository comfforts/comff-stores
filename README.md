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
- `cd pkg/utils/geohash && go test -v`
- `cd pkg/services/store && go test -v`
- `cd internal/server && go test -v`


- `gopkg.in/natefinch/lumberjack.v2`

- `brew install grpccurl`
- `grpcurl -plaintext localhost:8080 store.v1.Stores/GetStats`
- `grpcurl -d '{"postalCode": "92612"}' -plaintext localhost:8080 store.v1.Stores/GeoLocate`
- `grpcurl -d '{"postalCode": "92612", "distance": 5}' -plaintext localhost:8080 store.v1.Stores/SearchStore`
- `grpcurl -d '{"city": "Hong Kong", "name": "Plaza Hollywood", "country": "CN", "longitude": 114.20169067382812, "latitude":  22.340700149536133, "storeId": 1}' -plaintext localhost:8080 store.v1.Stores/AddStore`
- `grpcurl -d '{"storeId": 1}' -plaintext localhost:8080 store.v1.Stores/GetStore`
- errors
    - `Failed to dial target host "localhost:8080": tls: first record does not look like a TLS handshake`
        - Use -plaintext, not -insecure. The -insecure flag means that TLS is still used, but that it does not bother authenticating the remote server (hence it being insecure). If no TLS is in use at all, -plaintext is the flag you want.
        - `-connect-timeout` flag with `grpccurl`

- grpcurl -import-path api/v1 -proto store.proto list

- `cp cmd/store/config.example.json cmd/store/config.json`
- add google api key
- `go run store.go`