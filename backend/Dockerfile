# 第一阶段：构建阶段
FROM golang:1.25.1-alpine3.22 AS builder

WORKDIR /app
RUN set -eux; \
    for repo in "https://mirrors.aliyun.com/alpine" "https://dl-cdn.alpinelinux.org/alpine"; do \
        printf '%s/v3.22/main\n%s/v3.22/community\n' "$repo" "$repo" > /etc/apk/repositories; \
        apk add --no-cache git && break; \
        echo "apk add git failed on ${repo}, trying next repo"; \
    done; \
    command -v git

# 可通过 --build-arg GOPROXY=... 覆盖，默认提供多级回退
ARG GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}

COPY go.mod go.sum ./
RUN set -eux; \
        for i in 1 2 3 4 5; do \
            go mod download && break; \
            echo "go mod download failed, retry ${i}/5"; \
            if [ "$i" -eq 5 ]; then exit 1; fi; \
        done

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/myapp

# 第二阶段：运行阶段
FROM alpine:3.22

# 安装必要数据
RUN set -eux; \
    for repo in "https://mirrors.aliyun.com/alpine" "https://dl-cdn.alpinelinux.org/alpine"; do \
        printf '%s/v3.22/main\n%s/v3.22/community\n' "$repo" "$repo" > /etc/apk/repositories; \
        apk --no-cache add ca-certificates tzdata && break; \
        echo "apk add ca-certificates tzdata failed on ${repo}, trying next repo"; \
    done; \
    apk info -e ca-certificates tzdata
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