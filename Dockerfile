# ----------------------------
# 1. 构建阶段
# ----------------------------
FROM golang:1.25.6-alpine AS builder

# 安装 git 等必要工具
RUN apk add --no-cache git

# 设置工作目录
WORKDIR /app

COPY go.mod go.sum ./

# run with:
#     go mod tidy
#     go mod vendor
COPY vendor ./vendor
#RUN go mod download

COPY . .

#RUN go build -o diarygo ./cmd/diarygo/main.go
RUN go build -mod=vendor -o diarygo ./cmd/diarygo/main.go

# ----------------------------
FROM alpine:3.23

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/diarygo .

COPY --from=builder /app/web ./web

VOLUME ["/app/config", "/app/data"]

EXPOSE 8080

CMD ["./diarygo"]
