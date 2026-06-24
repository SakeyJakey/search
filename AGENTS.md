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

To update the search index, follow this sequence:
1.  **Crawl:** Run `crawler` to ingest HTML content into `raw_content`.
2.  **Post-process:** Run `postprocessing` to parse `raw_content`, compute embeddings/TF-IDF, update indices, and move data to `processed_content`.
3.  **Search:** Run `frontend` to query the populated search indices.

## 3. Applications

### 3.1. `frontend` (Web Server)

*   **Purpose:** Serves the web interface for the search engine, handling both vector-based and TF-IDF searches.
*   **Run Command:**
    ```bash
    go run frontend/main.go
    ```
*   **Access:** The server runs on `http://localhost:9091`.
*   **Assets:** Static files (CSS/JS) are served from `public/css` and `public/js`. HTML templates are in `public/tmpl/*.html`.

### 3.2. `crawler` (Web Crawler)

*   **Purpose:** Rapidly crawls URLs and stores raw HTML in the `raw_content` table.
*   **Run Command:**
    ```bash
    go run crawler/main.go [path/to/seeds.txt]
    ```
    *   **Performance Note:** Crawler is network-bound. It does *not* perform indexing.
    *   **Concurrency:** Uses `runtime.NumCPU()` goroutines.

### 3.3. `postprocessing` (Indexer)

*   **Purpose:** Processes `raw_content`, generates search indices (Vector + TF-IDF), and archives data to `processed_content`.
*   **Run Command:**
    ```bash
    go run postprocessing/main.go
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
