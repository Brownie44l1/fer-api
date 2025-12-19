FROM golang:1.23-bookworm AS builder

# Install ONNX Runtime
RUN apt-get update && apt-get install -y wget && \
    wget https://github.com/microsoft/onnxruntime/releases/download/v1.16.3/onnxruntime-linux-x64-1.16.3.tgz && \
    tar -xzf onnxruntime-linux-x64-1.16.3.tgz && \
    cp onnxruntime-linux-x64-1.16.3/lib/* /usr/local/lib/ && \
    cp -r onnxruntime-linux-x64-1.16.3/include/* /usr/local/include/ && \
    ldconfig && \
    rm -rf onnxruntime-linux-x64-1.16.3*

WORKDIR /app

# Copy dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=1 go build -o fer-api cmd/server/main.go

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy ONNX Runtime libs from builder
COPY --from=builder /usr/local/lib/libonnxruntime* /usr/local/lib/
RUN ldconfig

WORKDIR /app

# Copy binary and models
COPY --from=builder /app/fer-api .
COPY --from=builder /app/models ./models

EXPOSE 8080

CMD ["./fer-api"]