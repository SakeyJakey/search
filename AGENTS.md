# OpenCode Agent Instructions for Summative Project

This repository contains a Go project with three main applications: `frontend`, `crawler`, and `postprocessing`. All applications interact with a PostgreSQL database.

## 1. Project Setup

*   **Initialize Go Modules (if not already done):**
    ```bash
    go mod init summative # Or an appropriate module name
    go mod tidy
    ```
*   **Database:** This project requires a PostgreSQL database running at `postgres://localhost/search`. Ensure it is running and accessible before starting any application.
    *   Consider creating the database with `createdb search` if it doesn't exist.
    *   The `db` directories within each module contain SQL queries and models.

## 2. Operational Workflow (Crucial)

To update the search index and perform evaluations, follow this sequence:
1.  **Crawl:** Run `crawler` to ingest HTML content into `raw_content`.
    *   **Note:** The crawler automatically clears the database using `TRUNCATE ... CASCADE` on startup to ensure a clean state and uses a custom User-Agent.
2.  **Post-process:** Run `postprocessing` to parse `raw_content`, compute embeddings/TF-IDF, update indices, and move data to `processed_content`.
    *   **Note:** The post-processing tool automatically creates a database index on `token_id` to optimize performance and clears indices using `TRUNCATE ... CASCADE` on startup. Supports `-tfidf-only` flag to skip embeddings.
3.  **Search:** Run `api` to query the populated search indices.
4.  **Evaluate:** Navigate to `/evaluate` in the UI to perform blind tests. Results are persisted to the database.
5.  **Summarize:** Navigate to `/summary` in the UI to visualize precision metrics.

## 3. Applications

### 3.1. `api` (Web Server)

*   **Purpose:** Serves the API for the search engine, handling search queries, blind evaluations, and summaries.
*   **Run Command:**
    ```bash
    go run api/main.go [--no-embeddings]
    ```
*   **Access:** The API runs on `http://localhost:9091`.

### 3.2. `crawler` (Web Crawler)

*   **Purpose:** Rapidly crawls URLs and stores raw HTML/Titles in the `raw_content` and `urls` tables.
*   **Run Command:**
    ```bash
    go run crawler/main.go [path/to/seeds.txt]
    ```

### 3.3. `postprocessing` (Indexer)

*   **Purpose:** Processes `raw_content`, generates search indices (Vector + TF-IDF), and archives data to `processed_content`.
*   **Run Command:**
    ```bash
    go run postprocessing/main.go [-tfidf-only]
    ```

## 4. General Commands

*   **Build an executable:**
    ```bash
    go build -o bin/frontend frontend/main.go
    go build -o bin/crawler crawler/main.go
    go build -o bin/postprocessing postprocessing/main.go
    ```
    *   Executables will be created in the `bin/` directory.

## 5. Dependencies

*   Dependencies are managed implicitly through `go.mod` and `go.sum`.
*   Key libraries used: `github.com/jackc/pgx/v5`, `github.com/pgvector/pgvector-go`, `golang.org/x/net/html`.

## 6. Rules

*   Do not touch [the writeup document](Writeup.tex)
