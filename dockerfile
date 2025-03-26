# Stage 1: Build
FROM golang:1.23.5 AS builder

# Set environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Buat folder kerja di dalam container
WORKDIR /app

# Copy go.mod dan go.sum terlebih dahulu untuk caching dependency
COPY go.mod go.sum ./
RUN go mod tidy

# Copy seluruh project ke dalam container
COPY . .

# Compile aplikasi dengan entry point `cmd/main.go`
RUN go build -o main cmd/main.go

# Stage 2: Runtime
FROM alpine:latest

# Install dependencies yang diperlukan (misalnya untuk healthcheck)
RUN apk add --no-cache curl

# Buat folder kerja di dalam container
WORKDIR /root/

# Copy binary hasil build dari stage sebelumnya
COPY --from=builder /app/main .

# Expose port aplikasi
EXPOSE 8080

# Jalankan aplikasi
CMD ["./main"]
