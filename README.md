# URL Shortener

This is a simple URL shortener in Go, using Gin, SQLite and the GORM ORM to store the original URLs and the generated short codes.

## Features

- Shorten `http` and `https` URLs
- Look up the original URL of a short code
- Redirect short codes to their original URLs

## Requirements

- Go 1.23+
- A C compiler (`gcc`), because the SQLite driver uses cgo
- Docker (optional)

## How to run the project

1. Clone this repository.
2. Install the dependencies:

```bash
go mod download
```

3. Run the server:

```bash
go run .
```

The server runs on port `8080` and stores data in `./internal/database/database.db`.

For live reload during development, [air](https://github.com/air-verse/air) is configured in `.air.toml`:

```bash
air
```

### With Docker

```bash
docker compose up --build
```

The database is stored in the `data` volume, so it survives container restarts.

### Tests

```bash
go test ./...
```

## Configuration

| Variable   | Default                           | Description                                                                                                                   |
| ---------- | --------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| `PORT`     | `8080`                            | HTTP port.                                                                                                                    |
| `DB_PATH`  | `./internal/database/database.db` | SQLite file. The directory is created if missing. The Docker image uses `/data/database.db`.                                  |
| `BASE_URL` | empty                             | Public address used to build `short_url` (e.g. `https://sho.rt`). When empty, the request host is used, with `https` detected from TLS or the `X-Forwarded-Proto` header. |
| `GIN_MODE` | `debug`                           | Use `release` in production (already set in the Docker image).                                                               |

## Endpoints

| Method | Path                  | Description                            |
| ------ | --------------------- | -------------------------------------- |
| GET    | `/health`             | Health check.                          |
| POST   | `/shorten`            | Creates a short code for a URL.        |
| GET    | `/:shortURL`          | Returns the original URL as JSON.      |
| GET    | `/redirect/:shortURL` | Redirects (`302`) to the original URL. |

Errors use the shape `{"error": "message"}`.

**Shorten a URL**

```bash
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://go.dev"}'
# {"short_url":"http://localhost:8080/aZ3kP9"}
```

Returns `400` when the body is not JSON, the URL is missing, or it is not an absolute `http`/`https` URL.

**Get the original URL**

```bash
curl http://localhost:8080/aZ3kP9
# {"originalURL":"https://go.dev"}
```

**Redirect**

```bash
curl -i http://localhost:8080/redirect/aZ3kP9
# HTTP/1.1 302 Found
# Location: https://go.dev
```

Unknown codes return `404`.

### Postman

Import [`docs/postman/url-shortener.postman_collection.json`](docs/postman/url-shortener.postman_collection.json) into Postman. It documents every route with request and response examples and includes tests.

- Set the `baseUrl` collection variable to the server address (default `http://localhost:8080`).
- Run **Shorten URL** first: it saves the generated code in the `shortCode` variable used by the other requests.

## Deploy

### Render

The repository includes a [Render Blueprint](render.yaml) that creates a Docker web service with a 1 GB persistent disk mounted at `/data`.

1. In the Render dashboard, choose **New > Blueprint** and connect this repository.
2. Review the plan and confirm. Persistent disks require a paid instance type.
3. After the first deploy, the API is available at `https://<service-name>.onrender.com`.

Deploys run automatically after the CI checks pass on `master`. Optionally set `BASE_URL` in the service environment when the short links must use a different domain than the request host.

Limitations of a service with a disk: a single instance, and a few seconds of downtime on each deploy.

### Other platforms

The Docker image is self-contained:

1. Build from the `Dockerfile`.
2. Mount a persistent volume at `/data`. Without it, every redeploy erases the database.
3. Use `/health` as the health check path.

The container runs as root because platform volumes are often mounted owned by root, which a non-root user cannot write to.

SQLite supports a single instance only. Scaling to more than one instance requires moving to a server database such as PostgreSQL.
