# AGENTS Guide

This file defines mandatory rules for humans and coding agents contributing to this repository.

## 1. Project Mission

Build a blog indexing service used during static-site generation.

- The static-site pipeline uploads blog posts to this service.
- The service indexes post content for full-text search.
- The static site calls search APIs at runtime and renders results.

## 2. Tech Stack (Required)

- Language: Go 1.25+
- Database: MySQL 9
- API style: RESTful JSON API over HTTP
- Preferred logging library: `github.com/sirupsen/logrus`
- Preferred HTTP server library: `github.com/gin-gonic/gin`
- Preferred configuration format: TOML (`.toml`) for application settings (for example database and auth configuration).

Use modern language features and actively maintained libraries. Avoid deprecated APIs and abandoned packages.

## 3. Product Requirements

### 3.1 Ingestion and Indexing API

Provide an authenticated API for uploading/indexing a full post snapshot.

Top-level payload fields:

- `snapshot_id` (stable identifier for idempotent retries)
- `generated_at` (RFC3339)
- `posts` (full replacement list)

Minimum fields for each item in `posts`:

- `id` (stable external identifier)
- `title`
- `url`
- `content` (plain text or HTML source)
- `published_at` (RFC3339)

Rules:

- Only the owner/infrastructure pipeline is allowed to call this API.
- Requests must be authenticated and authorized before processing.
- Indexing must be idempotent by `snapshot_id` (safe retry behavior).
- Each ingestion request is a full replacement snapshot; posts not present in the snapshot must be removed from storage.
- Return clear, typed errors and HTTP status codes.

### 3.2 Search API

Provide a public (or selectively protected) search endpoint.

- Accept keyword query and pagination params.
- Execute full-text search against indexed posts.
- Return for each hit:
  - `title`
  - `url`
  - `snippet` (highlighted or context fragment)
  - optional scoring metadata (`score`, `matched_terms`)

## 4. Chinese Text Support (Non-negotiable)

The corpus includes Chinese content. Design must support Chinese tokenization and retrieval quality.

Required guidelines:

- Do not rely only on whitespace tokenization.
- Use a strategy compatible with MySQL 9 full-text capabilities (for example ngram parser) and document its limitations.
- Keep tokenizer/indexing behavior deterministic and testable.
- Generate snippets safely with Unicode-aware logic (rune-safe processing).

If an external segmenter is introduced, wrap it behind an interface so it can be replaced without touching core domain logic.

## 5. Architecture and Boundaries

Prefer high cohesion and low coupling.

Recommended layout:

- `main.go` for app entrypoint
- `data/` for data models and validation
- `service/` for business orchestration
- `adapter/http/` for REST handlers
- `adapter/storage/` for MySQL repositories
- `adapter/index/` for tokenizer/index pipeline integration
- `config/` for TOML config loading
- `app/` for bootstrap and dependency wiring

Rules:

- Data and service layers must not import transport or framework details.
- Handlers validate/parse I/O; services own business flow.
- Repositories hide SQL details behind interfaces.
- Keep `github.com/gin-gonic/gin` scoped to HTTP adapter packages; do not leak Gin-specific types into data or service layers.
- Keep configuration loading/parsing centralized in `config`.
- Keep functions focused and side effects explicit.

## 6. API and Security Standards

- Use context propagation (`context.Context`) for all request-scoped operations.
- Enforce request timeouts and database timeouts.
- Validate all input and return structured error responses.
- Authentication for ingestion APIs is mandatory (for example bearer token or signed request).
- Do not commit raw secrets in TOML files.
- Never log secrets, tokens, or raw sensitive payloads.
- Prepare for rate limiting and abuse control on search endpoints.

## 7. Database and Indexing Standards

- Manage schema via migrations under a dedicated directory (for example `migrations/`).
- Keep schema and indexes versioned and reproducible.
- Use explicit transaction boundaries where needed.
- Design indexes for both exact lookup (`id`, `url`) and full-text search fields (`title`, `content`).
- Benchmark and review query plans for search paths.

## 8. Code Quality Rules

- Repository language policy: all code, comments, docs, commit messages, and API examples must be in English.
- Follow idiomatic Go style and package naming.
- Import `github.com/sirupsen/logrus` with alias `log`, and call it as `log.xxx`.
- File ending style: keep at most one trailing blank line at end-of-file; do not leave multiple trailing blank lines.
- Prefer composition over inheritance-like coupling.
- Return wrapped errors with actionable context.
- Avoid global mutable state.
- Keep public APIs minimal and stable.

## 9. Testing Policy (Required)

Unit tests are mandatory for all business logic.

Minimum expectations:

- Table-driven tests for services and validators.
- Repository tests for SQL behavior (with test DB or reliable integration harness).
- HTTP handler tests covering auth, validation, and response codes.
- Search tests covering Chinese queries, mixed-language content, and snippet generation.
- Deterministic tests (no sleep-based flakiness, controlled time/randomness).

Quality gates (recommended):

- `go test ./...` passes locally and in CI.
- Race detector for concurrent components (`go test -race ./...`).
- Coverage focus on `data` and `service`.

## 10. Observability and Operations

- Add structured logs with request identifiers using `github.com/sirupsen/logrus`.
- Expose health and readiness checks.
- Emit metrics for ingestion latency, index size growth, and search latency.
- Capture enough diagnostics for production incidents without leaking sensitive data.

## 11. Contributor Do / Do Not

Do:

- Keep changes small, reversible, and well-tested.
- Add/adjust tests with every behavior change.
- Document API and schema changes in the same PR.
- Remove dead code (unused functions, branches, and stale helpers) as part of each change.

Do not:

- Introduce deprecated dependencies or APIs.
- Mix domain logic into HTTP handlers or SQL layers.
- Merge code without tests for new behavior.
- Commit credentials, tokens, or local environment secrets.

## 12. Definition of Done

A change is done only when:

1. Behavior is implemented and aligned with this guide.
2. Unit/integration tests are added and passing.
3. API/contracts and migration impacts are documented.
4. Security and Chinese search requirements remain satisfied.
