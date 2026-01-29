# ----------------------------
# 1. 构建阶段
# ----------------------------
FROM golang:1.25.6-alpine AS builder

# 安装 git 等必要工具
RUN apk add --no-cache git bash

# 设置工作目录
WORKDIR /app

# 先复制 go.mod 和 go.sum 下载依赖（提高缓存命中率）
COPY go.mod go.sum ./
RUN go mod download

# 复制项目源码
COPY . .

# 构建二进制文件
RUN go build -o diarygo ./cmd/diarygo/main.go

# ----------------------------
# 2. 运行阶段（轻量级）
# ----------------------------
FROM alpine:3.18

# 安装 CA 证书
RUN apk add --no-cache ca-certificates

# 设置工作目录
WORKDIR /app

# 复制编译好的二进制文件
COPY --from=builder /app/diarygo .

# 复制 web 目录里的静态文件和模板（打包进镜像）
COPY --from=builder /app/web ./web

# 配置目录和数据目录，容器里仍然放在 /app/config 和 /app/data
VOLUME ["/app/config", "/app/data"]

# 暴露端口
EXPOSE 8080

# 启动应用
CMD ["./diarygo"]
