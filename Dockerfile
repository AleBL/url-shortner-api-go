# syntax=docker/dockerfile:1

# go-sqlite3 needs cgo, so build on Debian and run on a glibc-based distroless image.
FROM golang:1.23-bookworm AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /out/url-shortener . \
    && mkdir -p /out/data

# Runs as root on purpose: hosting platforms (Render, Railway) may mount persistent
# disks owned by root, and a non-root user would fail to create the SQLite file.
FROM gcr.io/distroless/base-debian12
WORKDIR /app

COPY --from=build /out/url-shortener /app/url-shortener
COPY --from=build /out/data /data

ENV GIN_MODE=release \
    PORT=8080 \
    DB_PATH=/data/database.db

EXPOSE 8080
ENTRYPOINT ["/app/url-shortener"]
