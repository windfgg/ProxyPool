FROM golang:1.21.4-alpine as builder

WORKDIR /src
COPY . /src

RUN go mod download && \
    go mod vendor && \
    CGO_ENABLED=0 go build -v -o ProxyPool .

FROM docker:stable

COPY --from=builder /src/ProxyPool /ProxyPool
COPY entrypoint.sh /entrypoint.sh
EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]
