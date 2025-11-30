# Minimal Dockerfile using pre-built binary
# Build locally first: CGO_ENABLED=0 GOOS=linux go build -o replay-api ./cmd/rest-api/main.go
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy pre-built binary
COPY replay-api ./

# Create required directories
RUN mkdir -p /app/replay_files /app/coverage

# Set environment variable to increase stack size
ENV GODEBUG=stackguard=99999000000000

EXPOSE 8080
CMD ["./replay-api"]
