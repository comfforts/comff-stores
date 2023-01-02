# Comfforts store service

## Developer's notes
### Setup
#### start
- `cp cmd/store/config.example.json cmd/store/config.json`
- Update config with valid values
- Create and add GCP creds `cmd/store/creds/<file-name>.json`
- Create and add CA and server certs files `cmd/store/certs/<cert-file>`
- Create and add ACL model and policy files `cmd/store/policies/<file>`
- `make start-server` or `cd cmd/store && go run store.go`
#### apis
- `brew install grpccurl`
- Create and add CA and client certs files `cmd/client/certs/<cert-file>`
- `cd cmd/client`
- `grpcurl -import-path api/v1 -proto store.proto list`
- `grpcurl -plaintext localhost:50051 store.v1.Stores/GetStats`
- `grpcurl -key certs/client-key.pem -cert certs/client.pem -cacert certs/ca.pem localhost:50051 store.v1.Stores/GetStats`
- `grpcurl -d '{"postalCode": "92612"}' -plaintext localhost:50051 store.v1.Stores/GeoLocate`
- `grpcurl -key certs/client-key.pem -cert certs/client.pem -cacert certs/ca.pem -d '{"postalCode": "92612"}' localhost:50051 store.v1.Stores/GeoLocate`
- Create a bulk data file such as `starbucks.json`, either add file into `data/` folder or upload to cloud storage bucket with `data/` upload path
    - `grpcurl -d '{"fileName": "starbucks.json"}' -plaintext localhost:50051 store.v1.Stores/StoreUpload`
    - `grpcurl -key certs/client-key.pem -cert certs/client.pem -cacert certs/ca.pem -d '{"fileName": "starbucks.json"}' localhost:50051 store.v1.Stores/StoreUpload`
- `grpcurl -d '{"postalCode": "92612", "distance": 5}' -plaintext localhost:50051 store.v1.Stores/SearchStore`
- `grpcurl -key certs/client-key.pem -cert certs/client.pem -cacert certs/ca.pem -d '{"postalCode": "92612", "distance": 5}' localhost:50051 store.v1.Stores/SearchStore`
- `grpcurl -d '{"city": "Hong Kong", "name": "Plaza Hollywood", "country": "CN", "longitude": 114.20169067382812, "latitude":  22.340700149536133, "storeId": 1}' -plaintext localhost:50051 store.v1.Stores/AddStore`
- `grpcurl -key certs/client-key.pem -cert certs/client.pem -cacert certs/ca.pem -d '{"city": "Hong Kong", "name": "Plaza Hollywood", "country": "CN", "longitude": 114.20169067382812, "latitude":  22.340700149536133, "storeId": 1}' localhost:50051 store.v1.Stores/AddStore`
- `grpcurl -d '{"id": <store id>}' -plaintext localhost:50051 store.v1.Stores/GetStore`
- `grpcurl -key certs/client-key.pem -cert certs/client.pem -cacert certs/ca.pem -d '{"id": <store id>}' localhost:50051 store.v1.Stores/GetStore`
#### errors
- `Failed to dial target host "localhost:50051": tls: first record does not look like a TLS handshake`
    - Use -plaintext, not -insecure or specify certs. The -insecure flag means that TLS is still used, but that it does not bother authenticating the remote server (hence it being insecure). If no TLS is in use at all, -plaintext is the flag you want.
    - Use `-connect-timeout` flag with `grpccurl`

### Maintenance
#### Running tests
- Add `creds/<file-name>.json` and `test-config.json` into following:
    - `pkg/jobs`
    - `pkg/services/geocode`,
    - `pkg/services/filestorage`
- `make run-test`

#### Proto code generation
- Build relevant modules and update proto defs
- `make build-proto`
- Update server handlers in `internal/server`

#### Build artifacts
- `make build-exec`
- `make build-docker`
- `make run-docker`

### Notes
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
- `gopkg.in/natefinch/lumberjack.v2`
- `gitlab.com/xerra/common/vincenty`
- `github.com/stretchr/testify`