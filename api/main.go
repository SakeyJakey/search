package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"api/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

const DATABASE_URL = "postgres://localhost/search"
const EMBEDDINGS_MAX = 300

var conn *pgx.Conn
var queries *db.Queries

type SearchResult struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

func VectorSearch(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	embedding := vec32(vecNorm(embeddings_embed(query)))
	results, err := queries.VectorSearch(ctx, db.VectorSearchParams{
		QueryEmbedding: pgvector.NewVector(embedding),
		Limit:          int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	searchResults := make([]SearchResult, len(results))
	for i, r := range results {
		searchResults[i] = SearchResult{
			URL:   r.Url,
			Title: "Vector",
		}
	}

	return searchResults, nil
}

func TFIDFSearch(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	tokens := strings.Split(strings.ToLower(query), " ")
	if len(tokens) == 0 {
		return []SearchResult{}, nil
	}

	tokenRecords := make([]int64, 0, len(tokens))
	for _, token := range tokens {
		tokenRow, err := queries.GetTokenIds(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("token lookup failed: %w", err)
		}
		tokenRecords = append(tokenRecords, tokenRow.ID)
	}

	if len(tokenRecords) == 0 {
		return []SearchResult{}, nil
	}

	results, err := queries.TFIDFSearch(ctx, db.TFIDFSearchParams{
		Tokens: tokens,
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("tfidf search failed: %w", err)
	}

	searchResults := make([]SearchResult, len(results))
	for i, r := range results {
		searchResults[i] = SearchResult{
			URL:   r.Url,
			Title: "TF-IDF",
		}
	}

	return searchResults, nil
}

func init() {
	embeddings_init()
}

func main() {
	connPool, err := pgxpool.New(context.Background(), DATABASE_URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer connPool.Close()

	queries = db.New(connPool)

	http.HandleFunc("/api/search", handleSearchApi)
	http.HandleFunc("/api/evaluate", handleEvaluateApi)

	log.Println("API Server running on :9091")
	if err := http.ListenAndServe(":9091", nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func handleSearchApi(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Query string `json:"query"`
		Type  string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	fmt.Printf("Searched for %s", req.Query)

	var results []SearchResult
	var err error

	if req.Type == "tfidf" {
		results, err = TFIDFSearch(r.Context(), req.Query, 30)
	} else {
		results, err = VectorSearch(r.Context(), req.Query, 30)
	}

	if err != nil {
		log.Printf("Search error: %v", err)
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func handleEvaluateApi(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Query string `json:"query"`
		Type  string `json:"type"`
		Score int32  `json:"score"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := queries.SaveEvaluation(r.Context(), db.SaveEvaluationParams{
		Query:      req.Query,
		SearchType: req.Type,
		Score:      req.Score,
	})
	if err != nil {
		log.Printf("Save evaluation error: %v", err)
		http.Error(w, "Failed to save evaluation", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
