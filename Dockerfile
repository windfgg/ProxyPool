FROM golang:1.21.4-alpine as builder

WORKDIR /src
COPY . /src

RUN apk add --no-cache git

RUN rm -rf goproxy && git clone https://github.com/windfgg/goproxy.git

RUN go mod download && \
    go mod vendor && \
    CGO_ENABLED=0 go build -v -o ProxyPool .

FROM docker:stable

COPY --from=builder /src/ProxyPool /ProxyPool
EXPOSE 8080
ENTRYPOINT ["/ProxyPool"]
