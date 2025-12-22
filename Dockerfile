# Build stage
FROM golang:1.21-alpine AS builder
# Install gcc for SQLite wrok (CGO_ENABLED=1)
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY . .
RUN go mod tidy
# Build file executable
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# Run stage
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/static ./static
# Open port 8080
EXPOSE 8080
CMD ["./main"]