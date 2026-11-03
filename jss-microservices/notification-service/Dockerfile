###
# Notification service container definition.  This multi‑stage build
# compiles the Go binary in a minimal builder stage then copies the
# resulting binary into a tiny Alpine runtime.  The resulting image
# contains only the compiled application and a wait‑for script.
###

FROM golang:1.24.6-alpine AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd

FROM alpine:latest

WORKDIR /app
COPY --from=build /app/main /app/main
COPY --from=build /app/wait-for.sh /app/wait-for.sh
RUN chmod +x /app/wait-for.sh

EXPOSE 8082 50022

CMD ["/app/main"]