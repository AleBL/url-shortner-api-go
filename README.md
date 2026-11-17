# URL Shortener

This is a simple URL shortener in Go, using Gin, the GORM ORM and SQLite (or PostgreSQL) to store the original URLs and the generated short codes.

## Features

- Shorten `http` and `https` URLs
- Redirect short links to their original URLs
- Look up the original URL of a short code

## Requirements

- Go 1.25+
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

# against PostgreSQL
DATABASE_URL="postgres://postgres:test@localhost:5432/shortener?sslmode=disable" go test -p 1 ./...
```

## Configuration

| Variable   | Default                           | Description                                                                                                                   |
| ---------- | --------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| `PORT`     | `8080`                            | HTTP port.                                                                                                                    |
| `DATABASE_URL` | empty                         | PostgreSQL connection string (e.g. `postgres://user:pass@host/db?sslmode=require`). When set, it is used instead of SQLite. |
| `DB_PATH`  | `./internal/database/database.db` | SQLite file, used when `DATABASE_URL` is empty. The directory is created if missing. The Docker image uses `/data/database.db`. |
| `BASE_URL` | empty                             | Public address used to build `short_url` (e.g. `https://sho.rt`). When empty, the request host is used, with `https` detected from TLS or the `X-Forwarded-Proto` header. |
| `ALLOWED_ORIGINS` | empty                      | Comma-separated frontend origins allowed by CORS (e.g. `https://app.example.com,http://localhost:5173`). When empty, browsers only accept same-origin calls. |
| `RATE_LIMIT_PER_MINUTE` | `10`                   | Requests per minute per client IP on `POST /shorten`. `0` disables the limit.                                                 |
| `CLIENT_IP_HEADER` | empty                     | Header where the hosting proxy puts the real client IP (see [Rate limiting](#rate-limiting)). When empty, the TCP peer address is used. |
| `DEBUG_CLIENT_IP` | empty                      | `true` enables `GET /debug/client-ip`, used to find the right `CLIENT_IP_HEADER`. Keep it off otherwise.                     |
| `GIN_MODE` | `debug`                           | Use `release` in production (already set in the Docker image).                                                               |

## Endpoints

| Method | Path                  | Description                            |
| ------ | --------------------- | -------------------------------------- |
| GET    | `/health`             | Health check.                          |
| POST   | `/shorten`            | Creates a short code for a URL.        |
| GET    | `/:shortURL`          | Redirects (`302`) to the original URL. |
| GET    | `/api/urls/:shortURL` | Returns the original URL as JSON.      |

Errors use the shape `{"error": "message"}`. `POST /shorten` returns `429` with a `Retry-After` header when a client exceeds the rate limit.

**Shorten a URL**

```bash
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://go.dev"}'
# {"short_url":"http://localhost:8080/aZ3kP9"}
```

Returns `400` when the body is not JSON, the URL is missing, or it is not an absolute `http`/`https` URL.

**Redirect** (the `short_url` itself)

```bash
curl -i http://localhost:8080/aZ3kP9
# HTTP/1.1 302 Found
# Location: https://go.dev
```

**Get the original URL**

```bash
curl http://localhost:8080/api/urls/aZ3kP9
# {"originalURL":"https://go.dev"}
```

Unknown codes return `404`.

### Postman

Import [`docs/postman/url-shortener.postman_collection.json`](docs/postman/url-shortener.postman_collection.json) into Postman. It documents every route with request and response examples and includes tests.

- Set the `baseUrl` collection variable to the server address (default `http://localhost:8080`).
- Run **Shorten URL** first: it saves the generated code in the `shortCode` variable used by the other requests.

## Deploy

### Render (free)

The repository includes a [Render Blueprint](render.yaml) that creates a free Docker web service. Free services have no persistent disk and lose local files on every deploy, restart or spin-down, so the data is stored in an external PostgreSQL database. [Neon](https://neon.com) has a permanent free plan that does not require a credit card.

1. Create a Neon project in the region closest to Render's (AWS US East, N. Virginia) and copy its connection string. Prefer the direct connection (pooling disabled).
2. In the Render dashboard, choose **New > Blueprint** and connect this repository.
3. Paste the Neon connection string when Render asks for `DATABASE_URL`.
4. After the first deploy, the API is available at `https://<service-name>.onrender.com`.

The Blueprint sets `CLIENT_IP_HEADER=CF-Connecting-IP`: Render sits behind Cloudflare, which writes the real client IP to that header.

Deploys run automatically after the CI checks pass on `master`. Optionally set `BASE_URL` in the service environment when the short links must use a different domain than the request host.

Free plan limitations:

- Render spins the service down after 15 minutes without traffic; the next request takes about a minute.
- Render grants 750 free instance hours per month, enough for one service running all month.
- Neon's free plan has 1 GB of storage per project and suspends the database after 5 minutes idle (it resumes in under a second).

### Rate limiting

`POST /shorten` is limited per client IP. Behind a proxy, such as Render's, every request comes from the proxy address, so the API must read the client IP from a header the proxy sets. `X-Forwarded-For` is never trusted, because clients can forge it.

To find the right header on a new platform:

1. Set `DEBUG_CLIENT_IP=true` and redeploy.
2. Send forged values and compare the response with your real IP (`curl -s https://api.ipify.org`):

   ```bash
   curl -s https://<service>/debug/client-ip \
     -H "X-Forwarded-For: 1.2.3.4" -H "True-Client-IP: 1.2.3.4" -H "X-Real-IP: 1.2.3.4"
   ```

   Do not send `CF-Connecting-IP` to a platform behind Cloudflare: Cloudflare rejects the request with `error code: 1000`. If the response shows `Cf-Connecting-Ip`, Cloudflare set it.

3. A header that shows your real IP, even though you sent `1.2.3.4`, is overwritten by the proxy and safe to use. Set `CLIENT_IP_HEADER` to its name.
4. Remove `DEBUG_CLIENT_IP` and redeploy.

Until `CLIENT_IP_HEADER` is set, all clients behind the proxy share a single limit.

### Other platforms

The Docker image is self-contained:

1. Build from the `Dockerfile`.
2. Set `DATABASE_URL` to a PostgreSQL database, or mount a persistent volume at `/data` to use SQLite. Without either, every redeploy erases the database.
3. Use `/health` as the health check path.

The container runs as root because platform volumes are often mounted owned by root, which a non-root user cannot write to.

SQLite supports a single instance only. Scaling to more than one instance requires moving to a server database such as PostgreSQL.
