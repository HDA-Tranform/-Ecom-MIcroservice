# order-service/Dockerfile
# ---------- build stage ----------
FROM golang:1.24.6-alpine AS builder
WORKDIR /app

# copy go.mod/go.sum trước để cache deps
COPY go.mod go.sum ./
RUN go mod download

# copy phần còn lại
COPY . .

# build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /order-service ./cmd

# ---------- runtime stage ----------
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /order-service /app/order-service

# cổng gRPC và HTTP Gateway
EXPOSE 50033 8083

CMD ["/app/order-service"]
