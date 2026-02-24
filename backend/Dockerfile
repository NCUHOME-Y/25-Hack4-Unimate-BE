# 第一阶段：构建阶段
FROM golang:1.25.1-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/myapp

# 第二阶段：运行阶段
FROM alpine:latest

# 安装必要数据
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

# 创建非 root 用户提高安全性
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 关键：不要在这里 COPY .env，环境变量应由 Azure 注入
# 确保程序里读取 DB_DSN 等变量使用的是 os.Getenv()

RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 8080

CMD ["./main"]