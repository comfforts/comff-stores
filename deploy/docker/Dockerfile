FROM alpine:latest as builder

ARG BUILD_OS=linux
ARG BUILD_ARCH=arm64

WORKDIR /stores

RUN addgroup --system --gid 1001 comfforts
RUN adduser --system --uid 1001 store

COPY cmd/store/build/"$BUILD_OS"_"$BUILD_ARCH"/store ./
COPY cmd/store/config.json ./
COPY cmd/store/creds ./creds
COPY cmd/store/certs ./certs
COPY cmd/store/policies ./policies

RUN chown -R store /stores

USER store

EXPOSE 50051
ENV PORT 50051

CMD ["./store"]
