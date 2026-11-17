# syntax=docker/dockerfile:1

# go-sqlite3 needs cgo, so build on Debian and run on a glibc-based distroless image.
FROM golang:1.23-bookworm AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /out/url-shortener . \
    && mkdir -p /out/data

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app

COPY --from=build /out/url-shortener /app/url-shortener
# Pre-create the data dir owned by the nonroot user so a fresh volume is writable.
COPY --from=build --chown=65532:65532 /out/data /data

ENV GIN_MODE=release \
    PORT=8080 \
    DB_PATH=/data/database.db

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/url-shortener"]
