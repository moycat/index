# Index API Usage Guide

This document explains how to use the Index service API from frontend or tooling agents.

- Base URL: `https://index.moy.cat`
- API version prefix: `/v1`
- Content type: `application/json`
- Time format: RFC3339 (for example `2026-03-19T10:00:00Z`)

## 1. API Overview

The service has two main API groups:

1. Public search API for runtime search in your site/app.
2. Protected ingestion API for full snapshot indexing (replace-all behavior).

Current endpoints:

- `GET /healthz`
- `GET /readyz`
- `GET /v1/search`
- `PUT /v1/posts/snapshot` (requires Bearer token)

## 2. Common Conventions

### 2.1 Request Headers

Use these headers when applicable:

- `Content-Type: application/json` (for JSON request body)
- `Authorization: Bearer <token>` (required for ingestion endpoint)
- `X-Request-Id: <id>` (optional but recommended for tracing)

### 2.2 Error Shape

When request fails, API returns:

```json
{
  "error": {
    "code": "invalid_argument",
    "message": "invalid_argument: query is required"
  }
}
```

Error code mapping:

- `unauthorized` -> HTTP `401`
- `invalid_argument` -> HTTP `400`
- `internal` -> HTTP `500`

## 3. Health Endpoints

### 3.1 GET /healthz

Check process health.

Request:

```bash
curl -i https://index.moy.cat/healthz
```

Response `200`:

```json
{
  "status": "ok"
}
```

### 3.2 GET /readyz

Check readiness.

Request:

```bash
curl -i https://index.moy.cat/readyz
```

Response `200`:

```json
{
  "status": "ready"
}
```

## 4. Search API (Public)

### 4.1 GET /v1/search

Search indexed posts by keyword.

Query parameters:

- `q` (string, required): search keyword.
- `page` (int, optional): page number, default `1`, min effectively `1`.
- `page_size` (int, optional): default `10`, max `50`.

Important runtime behavior:

- If `q` is empty -> `400 invalid_argument`.
- If `page < 1` -> treated as `1`.
- If `page_size < 1` -> treated as `10`.
- If `page_size > 50` -> clamped to `50`.

Request examples:

```bash
curl -G 'https://index.moy.cat/v1/search' \
  --data-urlencode 'q=中文 检索' \
  --data-urlencode 'page=1' \
  --data-urlencode 'page_size=10'
```

```bash
curl -G 'https://index.moy.cat/v1/search' \
  --data-urlencode 'q=golang'
```

Success response `200`:

```json
{
  "query": "中文 检索",
  "hits": [
    {
      "title": "Go 和中文检索",
      "url": "https://example.com/1",
      "snippet": "...关于中文搜索能力和分词策略...",
      "score": 1.53,
      "matched_terms": ["中文", "检索"]
    }
  ]
}
```

Hit fields:

- `title` (string)
- `url` (string)
- `snippet` (string)
- `score` (number, optional)
- `matched_terms` (string array, optional)

### 4.2 Search Error Example

```bash
curl -G 'https://index.moy.cat/v1/search' --data-urlencode 'q='
```

Response `400`:

```json
{
  "error": {
    "code": "invalid_argument",
    "message": "invalid_argument: query is required"
  }
}
```

## 5. Snapshot Ingestion API (Protected)

### 5.1 PUT /v1/posts/snapshot

Upload a full snapshot of all posts. This is a replace-all operation.

Authentication:

- Required: `Authorization: Bearer <ingest-token>`

Top-level request fields:

- `snapshot_id` (string, required): stable idempotency key for this snapshot.
- `generated_at` (string RFC3339, required): snapshot generation time.
- `posts` (array, required; can be empty): full current post list.

Per-post fields:

- `id` (string, required)
- `title` (string, required)
- `url` (string, required, must be valid URL)
- `content` (string, required)
- `published_at` (string RFC3339, required)

Semantics:

- The uploaded `posts` is treated as the full source of truth.
- Service upserts all posts in this snapshot.
- Service deletes posts not present in this snapshot.
- Re-sending the same `snapshot_id` is idempotent.

Request example:

```bash
curl -X PUT 'https://index.moy.cat/v1/posts/snapshot' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <your-token>' \
  -d '{
    "snapshot_id": "snapshot-2026-03-19T10:00:00Z",
    "generated_at": "2026-03-19T10:00:00Z",
    "posts": [
      {
        "id": "post-1",
        "title": "Example",
        "url": "https://example.com/p/1",
        "content": "Post content...",
        "published_at": "2026-03-01T10:00:00Z"
      },
      {
        "id": "post-2",
        "title": "Another post",
        "url": "https://example.com/p/2",
        "content": "Second post content",
        "published_at": "2026-03-10T08:00:00Z"
      }
    ]
  }'
```

Success response `200`:

```json
{
  "status": "replaced",
  "snapshot_id": "snapshot-2026-03-19T10:00:00Z",
  "post_count": 2
}
```

### 5.2 Deletion by Snapshot

If you want to remove old posts, just omit them from the next snapshot.

If you want to clear all posts, send an empty snapshot list:

```bash
curl -X PUT 'https://index.moy.cat/v1/posts/snapshot' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <your-token>' \
  -d '{
    "snapshot_id": "snapshot-clear-2026-03-19T11:00:00Z",
    "generated_at": "2026-03-19T11:00:00Z",
    "posts": []
  }'
```

### 5.3 Ingestion Error Examples

Missing/invalid token:

```json
{
  "error": {
    "code": "unauthorized",
    "message": "unauthorized"
  }
}
```

Invalid time format:

```json
{
  "error": {
    "code": "invalid_argument",
    "message": "invalid_argument: generated_at must be RFC3339"
  }
}
```

Duplicate post ID in one snapshot:

```json
{
  "error": {
    "code": "invalid_argument",
    "message": "invalid_argument: duplicate post id: post-1"
  }
}
```

## 6. Frontend Integration Notes

### 6.1 Browser Search Example (fetch)

```js
const baseUrl = "https://index.moy.cat";

export async function searchPosts(query, page = 1, pageSize = 10) {
  const url = new URL(`${baseUrl}/v1/search`);
  url.searchParams.set("q", query);
  url.searchParams.set("page", String(page));
  url.searchParams.set("page_size", String(pageSize));

  const resp = await fetch(url.toString(), {
    method: "GET",
    headers: {
      "X-Request-Id": crypto.randomUUID(),
    },
  });

  const data = await resp.json();
  if (!resp.ok) {
    throw new Error(data?.error?.message || "search request failed");
  }
  return data;
}
```

### 6.2 Snapshot Publisher Example (Node/Agent)

```js
const baseUrl = "https://index.moy.cat";

export async function publishSnapshot(token, snapshot) {
  const resp = await fetch(`${baseUrl}/v1/posts/snapshot`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
      "X-Request-Id": crypto.randomUUID(),
    },
    body: JSON.stringify(snapshot),
  });

  const data = await resp.json();
  if (!resp.ok) {
    throw new Error(data?.error?.message || "snapshot publish failed");
  }
  return data;
}
```

Snapshot object shape:

```js
{
  snapshot_id: "snapshot-2026-03-19T10:00:00Z",
  generated_at: "2026-03-19T10:00:00Z",
  posts: [
    {
      id: "post-1",
      title: "Example",
      url: "https://example.com/p/1",
      content: "Post content...",
      published_at: "2026-03-01T10:00:00Z"
    }
  ]
}
```

## 7. Recommended Client-Side Strategies

- Always send a unique, stable `snapshot_id` for each generated snapshot.
- Keep ingestion token only in trusted server-side agents, never in browser frontend code.
- Validate/normalize URL and RFC3339 time before request to reduce `400` errors.
- For search UI, debounce input and cancel stale requests.
- Send `X-Request-Id` for easier troubleshooting with backend logs.

## 8. Quick Checklist for Frontend Agent

- Use base URL `https://index.moy.cat`.
- Use `GET /v1/search` for runtime search.
- Use `PUT /v1/posts/snapshot` with Bearer token for ingestion.
- Treat ingestion as full replacement, not incremental upsert.
- Handle `401`, `400`, and `500` with the documented error schema.

