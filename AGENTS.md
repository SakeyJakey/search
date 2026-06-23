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

## 2. Applications

### 2.1. `frontend` (Web Server)

*   **Purpose:** Serves the web interface for the search engine, handling both vector-based and TF-IDF searches.
*   **Run Command:**
    ```bash
    go run frontend/main.go
    ```
*   **Access:** The server runs on `http://localhost:9091`.
*   **Assets:** Static files (CSS/JS) are served from `public/css` and `public/js`. HTML templates are in `public/tmpl/*.html`.

### 2.2. `crawler` (Web Crawler)

*   **Purpose:** Crawls specified URLs, extracts content, and stores it (along with embeddings and TF-IDF tokens) in the PostgreSQL database.
*   **Run Command:**
    ```bash
    go run crawler/main.go [path/to/seeds.txt]
    ```
    *   **Seeds File:** By default, it looks for `seeds.txt` in the root. You can specify a different path as a command-line argument.
    *   **Concurrency:** Uses multiple goroutines for crawling.

### 2.3. `postprocessing` (TF-IDF Indexer)

*   **Purpose:** Computes and updates the TF-IDF index in the PostgreSQL database based on crawled data. This process clears and recomputes the `tf_idf_index` table.
*   **Run Command:**
    ```bash
    go run postprocessing/main.go
    ```

## 3. General Commands

*   **Build an executable:**
    ```bash
    go build -o bin/frontend frontend/main.go
    go build -o bin/crawler crawler/main.go
    go build -o bin/postprocessing postprocessing/main.go
    ```
    *   Executables will be created in the `bin/` directory (you may need to `mkdir bin`).

## 4. Dependencies

*   Dependencies are managed implicitly through `go.mod` and `go.sum` after `go mod init` and `go mod tidy`.
*   Key libraries used: `github.com/jackc/pgx/v5`, `github.com/pgvector/pgvector-go`, `golang.org/x/net/html`.
