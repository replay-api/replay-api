# Build stage
FROM golang:1.23-alpine AS build
WORKDIR /app
COPY cmd/minimal-api/main.go .
RUN CGO_ENABLED=0 go build -o replay-api main.go

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=build /app/replay-api ./
EXPOSE 8080
CMD ["./replay-api"]
