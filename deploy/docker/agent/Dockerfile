FROM alpine:latest as builder

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

RUN chown -R comffstore /comffstores

USER comffstore

EXPOSE 50051
ENV PORT 50051

CMD ["./comffstore", "--config-file", "config.yaml"]

