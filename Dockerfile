FROM golang:1.23-bookworm AS builder

WORKDIR /app

# Download ONNX Runtime 1.18.1 (supports API v18)
RUN wget https://github.com/microsoft/onnxruntime/releases/download/v1.18.1/onnxruntime-linux-x64-1.18.1.tgz && \
    tar -xzf onnxruntime-linux-x64-1.18.1.tgz && \
    mv onnxruntime-linux-x64-1.18.1 onnxruntime && \
    rm onnxruntime-linux-x64-1.18.1.tgz

# Install system ONNX Runtime for building
RUN cp onnxruntime/lib/* /usr/local/lib/ && \
    cp -r onnxruntime/include/* /usr/local/include/ && \
    ldconfig

# Copy dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build with CGO
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

# Copy ONNX Runtime libs to app directory
COPY --from=builder /app/onnxruntime/lib/ /app/lib/

# Create symlinks in /app/lib
RUN cd /app/lib && \
    ln -sf libonnxruntime.so.* libonnxruntime.so && \
    ln -sf libonnxruntime.so onnxruntime.so && \
    ls -la

# Copy binary and models
COPY --from=builder /app/fer-api .
COPY --from=builder /app/models ./models

# Set library path to look in /app/lib first
ENV LD_LIBRARY_PATH=/app/lib:/usr/local/lib:/usr/lib:$LD_LIBRARY_PATH

EXPOSE 8080

CMD ["./fer-api"]