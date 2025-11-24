# dependencies
# Use a local cache for the base image
FROM --platform=$BUILDPLATFORM golang:1.23 AS dependencies
WORKDIR /app 
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# build
FROM --platform=$BUILDPLATFORM golang:1.23 AS build
WORKDIR /app
COPY --from=dependencies /go/pkg /go/pkg
COPY . /app
RUN go mod tidy
RUN go mod download
RUN CGO_ENABLED=0 go build -o replay-api-http-service ./cmd/rest-api/main.go
RUN mkdir -p /app/replay_files
RUN mkdir -p /app/coverage

# runtime
FROM alpine:latest AS runtime
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=build /app/replay-api-http-service ./
COPY --from=build /app/replay_files ./replay_files
COPY --from=build /app/coverage ./coverage

# Set environment variable to increase stack size
ENV GODEBUG=stackguard=99999000000000

EXPOSE 4991
CMD ["./app/replay-api-http-service"]
