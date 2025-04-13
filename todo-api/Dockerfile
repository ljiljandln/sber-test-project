# Build
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o /todo-app ./cmd/main.go

# Final
FROM alpine:latest
COPY --from=builder /todo-app /todo-app
COPY --from=builder /app/migrations /migrations
EXPOSE 8081
CMD ["/todo-app"]