FROM golang:1.24.5-alpine AS builder
WORKDIR /app

# RUN go install github.com/air-verse/air@latest
# ENV PATH=$PATH:/go/bin

COPY go.* ./
RUN go mod download

COPY . .

# Build main service
RUN go build -o ./cmd/main ./cmd/main.go
RUN go build -o ./scripts/seed/seed ./scripts/seed/main.go

FROM migrate/migrate:v4.17.0 AS migrate

# HEALTHCHECK --interval=10s --timeout=3s --start-period=5s \
# CMD curl -f http://localhost:8001/health || exit 1

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/cmd/main /app/cmd/main
COPY --from=builder /app/scripts/seed/seed /app/scripts/seed/seed
COPY --from=builder /app/config.yaml /app/config.yaml
COPY --from=builder /app/scripts/seed/ /app/scripts/seed/
COPY --from=builder /app/db/ /app/db/
COPY --from=migrate /usr/local/bin/migrate /app/migrate

# COPY migrations ./migrations
# COPY entrypoint.sh /app/entrypoint.sh
# RUN chmod +x /app/entrypoint.sh

EXPOSE 8001 50001

# ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["./cmd/main"]
# CMD ["air", "-c", "./.air.toml"]
