# Comfforts Stores Service

`comff-stores` is the gRPC service that owns store records for Comfforts. It stores basic store identity and organization data, validates store addresses through the Comfforts Geo service, and supports location-aware search by address ID, address text, or latitude/longitude.

## Service Capabilities

The public API is defined in `api/stores/v1/stores.proto` as `stores.v1.Stores`.

| RPC | Product capability | Important behavior |
| --- | --- | --- |
| `AddStore` | Create a store for an organization. | Requires `org`, `name`, and `address_id`. The address ID is validated against Geo before the store is written. |
| `GetStore` | Fetch one store by ID. | Requires the MongoDB ObjectID returned by `AddStore`. |
| `UpdateStore` | Update store name, org, or address ID. | Requires store ID and at least one mutable field. |
| `DeleteStore` | Remove a store. | Requires store ID. |
| `SearchStore` | Find stores by organization, name, address ID, address string, or point. | Name/org searches are case-insensitive prefix matches. Address text and lat/lon are resolved through Geo. If a location is supplied without an explicit distance, the default radius is 5000 meters. |

The store model currently contains:

- `id`: MongoDB document ID.
- `name`: Store display name.
- `org`: Organization or tenant identifier.
- `address_id`: Geo address hash/ID.

## Architecture

The runtime entrypoint is `cmd/servers/stores/server.go`.

```text
gRPC client
  -> mTLS gRPC server
  -> auth/logging/metrics/tracing interceptors
  -> internal/delivery/stores/grpc_handler
  -> internal/usecase/services/stores
  -> internal/repo/stores
  -> MongoDB

stores service -> comff-geo-client -> Comfforts Geo service
```

Main implementation areas:

- `api/stores/v1/stores.proto`: public service contract.
- `internal/delivery/stores/grpc_handler`: gRPC handlers, auth, metadata logging, health, reflection, request metrics.
- `internal/usecase/services/stores`: business logic and Geo validation/geocoding.
- `internal/repo/stores`: MongoDB persistence and query behavior.
- `internal/infra/observability`: Prometheus metrics endpoint and OTLP tracing setup.
- `pkg/utils/environ`: environment-to-config helpers.
- `cmd/servers/stores/Dockerfile`: production and debug images.
- `k8s/stores`: Kind/Kubernetes deployment, service, config, policy, and cert-manager certificate resources.

## Business Rules

- A store must have `org`, `name`, and `address_id` when created.
- `AddStore` validates `address_id` with the Geo service before insertion.
- MongoDB enforces a unique index on `address_id`, so the current data model allows only one store per exact address ID.
- Search accepts any combination of `org`, `name`, and location fields, but at least one search parameter is required.
- If `SearchStore` receives `address_str`, the service asks Geo to geocode it and searches by the returned address hash.
- If `SearchStore` receives `latitude` and `longitude`, the service asks Geo to resolve that point and searches by the returned address hash.
- Distance search is implemented by truncating the Geo hash prefix before querying MongoDB. The response currently returns matched stores but does not populate per-store distance.

## Security And Authorization

The gRPC server runs with TLS configured from:
- `TLS_CA_FILE`
- `TLS_CERT_FILE`
- `TLS_KEY_FILE`

Policy lives in:

- `cmd/servers/stores/policies/policy.csv`
- `k8s/stores/stores-policy.yaml`

Supported policy actions are:

- `add-store`
- `get-store`
- `update-store`
- `delete-store`
- `search-stores`

## Dependencies

Runtime dependencies:

- MongoDB for store persistence.
- Comfforts Geo service through `github.com/comfforts/comff-geo-client`.
- cert-manager in Kubernetes for service certificates.
- OpenTelemetry collector for trace export when `OTEL_ENDPOINT` is set.
- Prometheus-compatible scraper for `/metrics`.

Key Go dependencies:

- `github.com/comfforts/comff-config`: TLS and authorization setup.
- `github.com/comfforts/comff-geo-client`: Geo validation and geocoding client.
- `github.com/comfforts/logger`: structured logging and trace-aware context fields.
- `google.golang.org/grpc`: gRPC server/client runtime.
- `go.mongodb.org/mongo-driver`: MongoDB driver.
- `go.opentelemetry.io/otel`: tracing and metrics.
- `github.com/prometheus/client_golang`: metrics HTTP handler.

## Configuration

Local env examples live in `env/stores.env`. Kubernetes config is generated into `k8s/stores/stores-config.yaml` and `k8s/stores/stores-secret.yaml`.

Important environment variables:

| Variable | Purpose |
| --- | --- |
| `SERVER_PORT` | gRPC port. Defaults to `62151`. |
| `METRICS_PORT` | Metrics HTTP port. Kubernetes uses `9467`. |
| `OTEL_ENDPOINT` | OTLP gRPC endpoint, for example `otel-collector.comff.svc.cluster.local:4317`. |
| `MONGO_PROTOCOL` | Mongo protocol, usually `mongodb`. |
| `MONGO_DBNAME` | Mongo database name. |
| `MONGO_HOST_NAME` | Direct Mongo host used by current startup path. |
| `MONGO_HOST_LIST` | Replica set host list. |
| `MONGO_DIR_CONN_PARAMS` | Direct Mongo connection params. |
| `MONGO_CLUS_CONN_PARAMS` | Replica set connection params. |
| `MONGO_USERNAME` / `MONGO_PASSWORD` | Mongo credentials. |
| `TLS_CA_FILE`, `TLS_CERT_FILE`, `TLS_KEY_FILE` | Server TLS files. |

Note: `server.go` currently calls `BuildMongoStoreConfig(true)`, so it uses `MONGO_HOST_NAME` and `MONGO_DIR_CONN_PARAMS`.

## Observability

The service exposes Prometheus metrics on `/metrics` at `METRICS_PORT`.

Current custom metric instruments are named with a `stores_` prefix 
- `stores_inflight_requests`
- `stores_requests_total`
- `stores_request_duration_seconds`

Each request metric includes:
- `rpc.method`
- `rpc.status`

The gRPC server also installs an OpenTelemetry gRPC stats handler. Service, usecase, and repository layers create spans such as:
- `stores.service.add`
- `stores.service.get`
- `stores.service.update`
- `stores.service.delete`
- `stores.service.search`
- `stores.repo.add`
- `stores.repo.search`

Logs include service, component, node, environment fields, RPC method/status/duration, peer address, certificate subject, and optional metadata:
- `x-request-id`
- `x-correlation-id`
- `user-agent`

## Calling The API

Install `grpcurl`:
```bash
brew install grpcurl
```

Port-forward the service:
```bash
kubectl -n comff port-forward svc/stores-server 62151:62151
```

Copy the generated Kubernetes client certs for local client use:
```bash
make cp-stores-cl-certs-k
```

List services with reflection:
```bash
grpcurl \
  -cert cmd/clients/stores/certs/client.pem \
  -key cmd/clients/stores/certs/client-key.pem \
  -cacert cmd/clients/stores/certs/ca.pem \
  localhost:62151 list
```

Example search:
```bash
grpcurl \
  -cert cmd/clients/stores/certs/client.pem \
  -key cmd/clients/stores/certs/client-key.pem \
  -cacert cmd/clients/stores/certs/ca.pem \
  -d '{"address_str":"92612","distance":5000}' \
  localhost:62151 stores.v1.Stores/SearchStore
```

If you see `tls: first record does not look like a TLS handshake`, the client and server disagree about TLS. Use valid cert flags for TLS, or use `-plaintext` only against a plaintext server.

## Maintaining The Service

When changing API capabilities:

1. Update `api/stores/v1/stores.proto`.
2. Run `make build-proto`.
3. Update mappings in `internal/domain/stores`.
4. Update handlers in `internal/delivery/stores/grpc_handler`.
5. Update business logic in `internal/usecase/services/stores`.
6. Update persistence behavior and indexes in `internal/repo/stores` if the data model changes.
7. Add or update tests with `make run-test`.
8. Update this README if product capability, dependencies, auth, or operational behavior changes.

When changing authorization:

1. Add or rename action constants in the gRPC handler.
2. Update `cmd/servers/stores/policies/policy.csv`.
3. Regenerate or update `k8s/stores/stores-policy.yaml`.
4. Apply with `kubectl apply -k k8s/stores`.

When changing deployment config:

1. Update `env/stores-config.env` or `env/stores-secret.env`.
2. Run `make build-stores-creds-k`.
3. Confirm `k8s/stores/kustomization.yaml` includes the generated resources.
4. Apply with `make set-stores-k`.

## Known Implementation Notes

- `UpdateStore` does not currently revalidate a changed `address_id` with Geo.
- `UpdateStoreResponse.store`, `SearchStoreResponse.geo`, and `StoreGeo.distance` are defined in the proto but are not currently populated by handlers.
- Handler errors are mostly returned as `Internal` after the service layer, even for domain cases such as missing store or duplicate store.
- The deployment has no explicit readiness or liveness probes yet.
