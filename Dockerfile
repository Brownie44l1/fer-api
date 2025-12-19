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

# Copy ONNX Runtime libraries
COPY --from=builder /usr/local/lib/libonnxruntime.so* /usr/local/lib/

# Create the symlink that yalue/onnxruntime_go expects
RUN cd /usr/local/lib && \
    ln -sf libonnxruntime.so.* libonnxruntime.so && \
    ln -sf libonnxruntime.so onnxruntime.so && \
    ldconfig && \
    ls -la /usr/local/lib/

# Copy binary and models
COPY --from=builder /app/fer-api .
COPY --from=builder /app/models ./models

# Set library path
ENV LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH

EXPOSE 8080

CMD ["./fer-api"]