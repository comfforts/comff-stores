# Comfforts store service

## Developer's notes
### Setup
#### start
- Setup mongo, metrics env vars
- Setup server certs - `cmd/servers/stores/certs/local-certs/`
- Setup ACL model and policy files `cmd/servers/stores/policies/<file>`
- `make start-server`
#### apis
- `brew install grpccurl`
- Create and add CA and client certs files `cmd/clients/stores/certs/local-certs/<cert-file>`
- `cd cmd/clients/stores`
- `grpcurl -import-path api/v1 -proto stores.proto list`
- `grpcurl -key certs/local-certs/client-key.pem -cert certs/local-certs/client.pem -cacert certs/local-certs/ca.pem -d '{"address_str": "92612", "distance": 5}' localhost:62151 stores.v1.Stores/SearchStores`
#### errors
- `Failed to dial target host "localhost:62151": tls: first record does not look like a TLS handshake`
    - Use -plaintext, not -insecure or specify certs. The -insecure flag means that TLS is still used, but that it does not bother authenticating the remote server (hence it being insecure). If no TLS is in use at all, -plaintext is the flag you want.
    - Use `-connect-timeout` flag with `grpccurl`

### Maintenance

#### Proto code generation
- Build relevant modules and update proto defs
- `make build-proto`
- Update server handlers in `internal/delivery/grpc_handler/`
