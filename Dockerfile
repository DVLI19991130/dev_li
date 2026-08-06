# 构建阶段
FROM golang:1.25.4 AS builder

WORKDIR /build

# 先复制依赖文件并下载，利用 Docker 缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 构建 Linux amd64 二进制 (CGO_ENABLED=0 静态编译)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -gcflags=all=-trimpath=/go \
    -asmflags=all=-trimpath=/go \
    -ldflags "-w -s" \
    -o mock .

# 运行阶段
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
RUN apk add --no-cache busybox-extras

RUN addgroup -g 1000 mock && \
    adduser -u 1000 -G mock -s /bin/sh -D mock

WORKDIR /app

# 从构建阶段复制二进制
COPY --from=builder /build/mock .

RUN mkdir -p /app/logs && chmod 777 /app/logs

USER mock

CMD ["sh", "-c", "cd /app && ./mock run -f mock.json"]
