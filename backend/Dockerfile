FROM golang:1.25.1-alpine

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main ./cmd/myapp

EXPOSE 8080

CMD ["./main"]