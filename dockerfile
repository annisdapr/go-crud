# # Stage 1: Build
# FROM golang:1.23.5 AS builder

# # Set environment variables
# ENV GO111MODULE=on \
#     CGO_ENABLED=0 \
#     GOOS=linux \
#     GOARCH=amd64

# # Buat folder kerja di dalam container
# WORKDIR /app

# # Copy go.mod dan go.sum terlebih dahulu untuk caching dependency
# COPY go.mod go.sum ./
# RUN go mod tidy

# # Copy seluruh project ke dalam container
# COPY . .

# # Compile aplikasi dengan entry point `cmd/main.go`
# RUN go build -o main cmd/main.go

# # Stage 2: Runtime
# FROM alpine:latest

# # Install dependencies yang diperlukan (misalnya untuk healthcheck)
# RUN apk add --no-cache curl

# # Buat folder kerja di dalam container
# WORKDIR /root/

# # Copy binary hasil build dari stage sebelumnya
# COPY --from=builder /app/main .

# # Expose port aplikasi
# EXPOSE 8080

# # Jalankan aplikasi
# CMD ["./main"]

# Stage 1: Build
FROM golang:1.23.5 AS builder

# Enable CGO dan set target Linux AMD64
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

# Install librdkafka-dev untuk Kafka bindings
RUN apt-get update && apt-get install -y librdkafka-dev

WORKDIR /app

# Copy module files dan download dependency
COPY go.mod go.sum ./
RUN go mod tidy

# Copy semua source code ke image
COPY . .

# Compile aplikasi
RUN go build -o main cmd/main.go

# Stage 2: Runtime
FROM debian:bullseye-slim

# Install librdkafka runtime dan curl
RUN apt-get update && apt-get install -y librdkafka1 curl && apt-get clean

WORKDIR /root/
COPY --from=builder /app/main .

# Expose port aplikasi
EXPOSE 8080

# Jalankan aplikasi
CMD ["./main"]
