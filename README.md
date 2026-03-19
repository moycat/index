# index

A blog indexing service for static-site generation.

## Features

- Authenticated full-snapshot ingestion API that replaces all indexed posts atomically.
- Full-text search API returning `title`, `url`, `snippet`, and score metadata.
- Chinese-aware tokenization for query term extraction (`2-gram` tokenizer in app layer).
- MySQL full-text index with `ngram` parser for Chinese retrieval support.
- TOML-based configuration as the only runtime configuration source.

## Tech Stack

- Go 1.25+
- MySQL 9
- HTTP: `github.com/gin-gonic/gin`
- Logging: `github.com/sirupsen/logrus`
- Config parser: `github.com/pelletier/go-toml/v2`

## Project Layout

- `main.go`: runnable API entrypoint.
- `data`: data models and validation.
- `service`: business services and orchestration.
- `adapter/http`: Gin router, handlers, auth middleware.
- `adapter/storage/mysql`: MySQL repository implementation.
- `adapter/index`: tokenizer and rune-safe snippet builder.
- `config`: TOML config loading and validation.
- `app`: dependency wiring and server bootstrap.
- `migrations`: SQL schema migrations.

## Configuration

Example config: `config.example.toml`

Create runtime config from the example file:

```bash
cp config.example.toml config.toml
```

Runtime config is file-only. Keep real secrets in your local `config.toml` and never commit them.

## Migration

Apply `migrations/001_init_posts.sql` to your MySQL database before running the service.

## Run

```bash
go run .
```

The default startup config path is `config.toml`. If the file is missing or incomplete, startup fails.

## API

### Health

- `GET /healthz`
- `GET /readyz`

### Ingest (Authenticated)

`PUT /v1/posts/snapshot`

Headers:

- `Authorization: Bearer <token>`

Body:

```json
{
  "snapshot_id": "snapshot-2026-03-19",
  "generated_at": "2026-03-19T10:00:00Z",
  "posts": [
    {
      "id": "post-1",
      "title": "Example",
      "url": "https://example.com/p/1",
      "content": "Post content...",
      "published_at": "2026-03-19T10:00:00Z"
    }
  ]
}
```

Ingestion semantics:

- Each call uploads a complete snapshot of all current posts.
- The service upserts all uploaded posts and deletes posts missing from the snapshot.
- Repeating the same `snapshot_id` is idempotent.

### Search

`GET /v1/search?q=keyword&page=1&page_size=10`

Response item fields:

- `title`
- `url`
- `snippet`
- `score` (optional)
- `matched_terms` (optional)

## Tests

```bash
go test ./...
```

```bash
go test -race ./...
```

## Notes on Chinese Search

- Database layer uses MySQL full-text index with `WITH PARSER ngram`.
- Application layer uses deterministic `2-gram` tokenization for query term extraction and matched-term/snippet generation.
- Snippet logic is rune-safe to avoid breaking Unicode text boundaries.
