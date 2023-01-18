FROM golang:1.19-alpine3.17 as build

RUN GRPC_HEALTH_PROBE_VERSION=v0.3.2 && wget -qO/go/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && chmod +x /go/bin/grpc_health_probe

FROM alpine:3.17

ARG BUILD_OS=linux
ARG BUILD_ARCH=arm64

WORKDIR /comffstores

RUN addgroup --system --gid 1001 comfforts
RUN adduser --system --uid 1002 comffstore

COPY cmd/cli/build/"$BUILD_OS"_"$BUILD_ARCH"/comffstore ./
COPY cmd/cli/config.json ./
COPY cmd/cli/creds ./creds
COPY cmd/cli/certs ./certs
COPY cmd/cli/policies ./policies

COPY --from=build /go/bin/grpc_health_probe /bin/grpc_health_probe

RUN chown -R comffstore /comffstores

USER comffstore

EXPOSE 50051
ENV PORT 50051

ENTRYPOINT ["./comffstore", "--data-dir", "data", "--bootstrap", "true"]

