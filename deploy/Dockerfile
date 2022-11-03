FROM  ccr.ccs.tencentyun.com/hyy-yu/alpine-timezone:3.15

ARG serviceName
WORKDIR /app

# Copy the binary named service from the context into our
# container image
# you can run : CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o service ./cmd/main.go
# generate a service binary
COPY service /app/service
COPY cmd/cfg.toml /app/cfg.toml
RUN mkdir -p ./conf

ENTRYPOINT ["./service"]