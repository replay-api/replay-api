# Minimal Dockerfile using pre-built binary
# Build locally first: CGO_ENABLED=0 GOOS=linux go build -o replay-api ./cmd/rest-api/main.go
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

# SECURITY: Run as non-root user
RUN addgroup -g 10001 -S appgroup && \
    adduser -u 10001 -S appuser -G appgroup

WORKDIR /app

# Copy pre-built binary
COPY replay-api ./

# Create required directories with correct ownership
RUN mkdir -p /app/replay_files /app/coverage /tmp && \
    chown -R appuser:appgroup /app /tmp

# SECURITY: Switch to non-root user
USER appuser

EXPOSE 8080
CMD ["./replay-api"]
