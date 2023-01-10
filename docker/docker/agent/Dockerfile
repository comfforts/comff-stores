FROM golang:1.19-alpine as build
WORKDIR /go/src/comffstores
COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/comffstores ./cmd/cli/comffstores

FROM scratch
COPY --from=build /go/bin/comffstores /bin/comffstores
ENTRYPOINT [ "/bin/comffstores" ]
