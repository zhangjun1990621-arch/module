# ---- 构建阶段 ----
FROM golang:1.24-alpine AS builder

# 配置国内 Go 模块代理（解决中国大陆无法访问 proxy.golang.org 的问题）
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

WORKDIR /build

# 先复制依赖文件，利用 Docker 层缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 静态编译
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/server ./cmd/server

# ---- 运行阶段 ----
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 拷贝编译产物与配置
COPY --from=builder /build/server .
COPY --from=builder /build/config ./config
COPY --from=builder /build/migrations ./migrations

EXPOSE 8090

ENTRYPOINT ["./server"]
