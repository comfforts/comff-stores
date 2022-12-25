FROM golang:1.19-alpine3.17 as builder

ARG BUILD_OS=linux
ARG BUILD_ARCH=arm64

WORKDIR /stores

RUN addgroup --system --gid 1001 comfforts
RUN adduser --system --uid 1001 store

COPY cmd/store/build/"$BUILD_OS"_"$BUILD_ARCH"/store ./
COPY cmd/store/config.json ./
COPY cmd/store/creds ./creds

# RUN chown -R store /stores
# RUN chown -R store ./creds/
# RUN chown -R store ./config.json

RUN chown -R store /stores ./creds ./config.json

USER store

EXPOSE 50051
ENV PORT 50051

CMD ["./store"]
