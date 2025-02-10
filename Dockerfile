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
COPY . /app
# ENV DEV_ENV docker
RUN CGO_ENABLED=0 go build -o replay-api-http-service ./cmd/rest-api/main.go
RUN mkdir -p /app/replay_files
RUN mkdir -p /app/coverage
RUN chown -R ${DEV_ENV}:${DEV_ENV} /app/replay_files
RUN chown -R ${DEV_ENV}:${DEV_ENV} /app/coverage

# runtime
FROM scratch AS runtime
COPY --from=build /app/replay-api-http-service ./app/
COPY --from=build /app/coverage ./app/coverage

# Set environment variable to increase stack size
ENV GODEBUG=stackguard=99999000000000

USER ${DEV_ENV}

EXPOSE 4991
CMD ["./app/replay-api-http-service"]
