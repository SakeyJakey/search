package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"runtime"
	"sync"

	"postprocessing/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

const DATABASE_URL = "postgres://localhost/search"
const EMBEDDINGS_MAX = 300

var conn *pgxpool.Pool
var queries *db.Queries
var tfidfOnly bool

func main() {
	flag.BoolVar(&tfidfOnly, "tfidf-only", false, "Only process TF-IDF indices, skip vector embeddings")
	flag.Parse()

	// conn, err := pgx.Connect(context.Background(), DATABASE_URL)
	var err error
	conn, err = pgxpool.New(context.Background(), DATABASE_URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	queries = db.New(conn)

	ctx := context.Background()
	
	fmt.Printf("[0/4] [  ] Clearing indices")
	query := "TRUNCATE TABLE tf_idf_index, tf_idf_counts RESTART IDENTITY CASCADE;"
	if !tfidfOnly {
		query = "TRUNCATE TABLE tf_idf_index, tf_idf_counts, vector_index RESTART IDENTITY CASCADE;"
	}
	_, err = conn.Exec(ctx, query)
	if err != nil {
		fmt.Printf("\r[0/4] [✘ ] Failed to clear tables: %v\n", err)
		return
	}

	fmt.Printf("\r[0/4] [🗸 ] Cleared indices\n")

	process()

	documentCount, err := queries.GetDocumentCount(ctx)
	if err != nil {
		fmt.Printf("\r[1/4] [✘ ] Failed to count documents: %v\n", err)
		return
	}

	if documentCount <= 0 {
		fmt.Printf("\r[1/4] [✘ ] No documents available to index.\n")
		return
	}

	fmt.Printf("\r[1/4] [🗸 ] Counted documents\n")

	fmt.Printf("[2/4] [  ] Fetching token IDs")
	tokens, err := queries.GetTokenIDs(ctx)
	if err != nil {
		fmt.Printf("\r[2/4] [✘ ] Failed to fetch token IDs\n")
		return
	}

	fmt.Printf("\r[2/4] [🗸 ] Fetched token IDs\n")

	// Parallel processing
	numWorkers := runtime.NumCPU()
	jobs := make(chan int64, numWorkers*2)
	var wg sync.WaitGroup

	// Start worker pool
	for range numWorkers {
		wg.Go(func() {
			for id := range jobs {
				localCtx := context.Background()
				
				df, err := queries.GetDocumentFrequencyByToken(localCtx, id)
				if err != nil {
					fmt.Fprintf(os.Stderr, "\nFailed to compute DF for token %d: %v\n", id, err)
					continue
				}
				
				if df <= 0 {
					continue
				}

				idf := math.Log(float64(documentCount) / float64(df))

				counts, err := queries.GetCountsForToken(localCtx, id)
				if err != nil {
					fmt.Fprintf(os.Stderr, "\nFailed to compute counts for token %d: %v\n", id, err)
					continue
				}

				for _, count := range counts {
					tf := float64(count.Count)
					tfidf := tf * idf
					if tfidf < 0 {
						tfidf = 0
					}

					err := queries.InsertTfIdfBatch(localCtx, db.InsertTfIdfBatchParams{
						UrlID:		count.UrlID,
						TokenID:	id,
						TfIdf:		tfidf,
					})
					
					if err != nil {
						fmt.Fprintf(os.Stderr, "\nFailed to insert for token ID %d: %v\n", id, err)
					}
				}
			}
		})
	}

	// Send jobs
	for _, id := range tokens {
		jobs <- id
	}
	close(jobs)
	wg.Wait()

	fmt.Printf("\r[3/4] [🗸 ] Computed TF-IDF\n")
	fmt.Printf("[4/4] [🗸 ] Completed post-processing\n")
}

func process() {
	ctx := context.Background()
	// 1. Get raw content
	rows, err := queries.GetUnprocessedRawContent(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get raw content: %v\n", err)
		return
	}

	if !tfidfOnly {
		embeddings_init()
	}

	for _, row := range rows {
		fmt.Printf("Processing URL ID: %d\n", row.UrlID)
		// 2. Index
		tfidf_add_token_index(ctx, row.UrlID, row.Content)
		if !tfidfOnly {
			embeddings_add(ctx, row.UrlID, row.Content)
		}

		// 3. Move to processed
		err = queries.MoveToProcessed(ctx, row.UrlID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to move to processed: %v\n", err)
			continue
		}

		// 4. Delete from raw
		err = queries.DeleteRawContent(ctx, row.UrlID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete from raw: %v\n", err)
		}
	}
}
