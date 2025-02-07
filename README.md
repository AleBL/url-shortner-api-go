# URL Shortener

This is a simple URL shortener in Go, using SQLite and the GORM ORM to store the original URLs and the generated shortcodes.

## Features

- Shorten URLs
- Redirect short URLs to their original versions

## How to run the project

1. Clone this repository.
2. Install the dependencies:

```bash
go mod tidy
```

3. Run the server:

```bash
go run main.go
```

4. The server will be running on port `8080`.

### Endpoints

- **POST /shorten**: Shortens a URL. Request example:

```bash
curl -X POST -d "url=http://example.com" http://localhost:8080/shorten
```

- **GET /{shortCode}**: Show to the original URL. Example:
