CREATE TABLE IF NOT EXISTS posts (
    id VARCHAR(128) NOT NULL,
    title VARCHAR(512) NOT NULL,
    url VARCHAR(750) NOT NULL,
    content LONGTEXT NOT NULL,
    published_at DATETIME(6) NOT NULL,
    snapshot_id VARCHAR(128) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_posts_url (url),
    KEY idx_posts_snapshot_id (snapshot_id),
    FULLTEXT KEY ft_posts_title_content (title, content) WITH PARSER ngram
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS ingest_snapshots (
    snapshot_id VARCHAR(128) NOT NULL,
    generated_at DATETIME(6) NOT NULL,
    post_count INT NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (snapshot_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
