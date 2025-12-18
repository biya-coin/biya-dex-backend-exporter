## biya-dex-backend-exporter
## - single stage build (先跑起来，后续再做多阶段/瘦身优化)

FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./
# 注意：`GOPROXY=... && go ...` 不会把 GOPROXY 导出到子进程，go 会继续使用默认 proxy.golang.org
RUN GOPROXY=https://goproxy.cn,direct GOSUMDB=off go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
  -o /usr/local/bin/biya-exporter ./cmd/exporter

EXPOSE 18080

# 默认不传 config，走内置 Default()；如需自定义请在 compose/启动参数中传 -config
ENTRYPOINT ["biya-exporter"]
