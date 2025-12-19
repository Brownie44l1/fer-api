FROM golang:1.23-bookworm AS builder

# Install ONNX Runtime
RUN apt-get update && apt-get install -y wget && \
    wget https://github.com/microsoft/onnxruntime/releases/download/v1.16.3/onnxruntime-linux-x64-1.16.3.tgz && \
    tar -xzf onnxruntime-linux-x64-1.16.3.tgz && \
    mkdir -p /usr/local/lib && \
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

# Build with CGO and specify library path
ENV CGO_ENABLED=1
ENV CGO_LDFLAGS="-L/usr/local/lib"
RUN go build -o fer-api cmd/server/main.go

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libgomp1 \
    libstdc++6 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the entire lib directory to ensure all dependencies
COPY --from=builder /usr/local/lib/libonnxruntime* /usr/lib/
COPY --from=builder /usr/local/lib/libonnxruntime* /usr/local/lib/

# Update library cache
RUN ldconfig

# Copy binary and models
COPY --from=builder /app/fer-api .
COPY --from=builder /app/models ./models

# Set multiple library paths
ENV LD_LIBRARY_PATH=/usr/lib:/usr/local/lib:/lib/x86_64-linux-gnu:/usr/lib/x86_64-linux-gnu

EXPOSE 8080

CMD ["./fer-api"]