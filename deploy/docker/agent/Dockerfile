FROM golang:1.19-alpine3.17 as build

WORKDIR /go/src/comffstore
COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/comffstore ./cmd/cli/comffstore.go

RUN GRPC_HEALTH_PROBE_VERSION=v0.3.2 && wget -qO/go/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && chmod +x /go/bin/grpc_health_probe

FROM alpine:3.17
# FROM ubuntu

WORKDIR /comffstores

# RUN echo '{"services":{"cloud_storage":{"creds_path":"creds/comfforts-prod.json"}},"geo_code":{"geocoder_key":"AIzaSyDY90JeSdXlsVyqgGjktHPGmpPZEE4iY7A","cached":true,"bucket_name":"comfforts-cache"},"jobs":{"store_loader":{"bucket_name":"comfforts-data"}}}' > config.json 

COPY --from=build /go/bin/comffstore ./
COPY --from=build /go/src/comffstore/cmd/cli/config.json ./

COPY --from=build /go/src/comffstore/cmd/cli/certs ./certs
COPY --from=build /go/src/comffstore/cmd/cli/creds ./creds
COPY --from=build /go/src/comffstore/cmd/cli/policies ./policies
COPY --from=build /go/src/comffstore/cmd/cli/data/geo ./data/geo

COPY --from=build /go/bin/grpc_health_probe /bin/grpc_health_probe

EXPOSE 50051

ENTRYPOINT ["./comffstore", "--data-dir", "data", "--bootstrap", "true"]
