package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func tfidf_count_tokens(text string) map[string]int32 {
	index := make(map[string]int32)
	for token := range strings.FieldsSeq(strings.ToLower(text)) {
		if len(token) <= 256 {
			index[token]++
		}
	}

	return index
}

func tfidf_add_token_index(ctx context.Context, urlID int64, content string) {
	index := tfidf_count_tokens(content)
	if len(index) == 0 {
		return
	}

	// Build stage rows
	tokenRows := make([][]any, 0, len(index))
	countRows := make([][]any, 0, len(index))
	for token, cnt := range index {
		tokenRows = append(tokenRows, []any{token})
		countRows = append(countRows, []any{urlID, token, int32(cnt)})
	}

	// Retry loop for deadlocks
	for i := 0; i < 3; i++ {
		err := func() error {
			tx, err := conn.Begin(ctx)
			if err != nil {
				return err
			}
			defer tx.Rollback(ctx)

			// Create temp tables
			_, err = tx.Exec(ctx, `
				CREATE TEMP TABLE IF NOT EXISTS tf_idf_tokens_stage (token text not null);
				CREATE TEMP TABLE IF NOT EXISTS tf_idf_counts_stage (
					url_id bigint not null,
					token text not null,
					"count" int not null
				);
			`)
			if err != nil {
				return err
			}

			// 1) COPY tokens into staging
			_, err = tx.CopyFrom(
				ctx,
				pgx.Identifier{"tf_idf_tokens_stage"},
				[]string{"token"},
				pgx.CopyFromRows(tokenRows),
			)
			if err != nil {
				return err
			}

			// 2) COPY counts into staging
			_, err = tx.CopyFrom(
				ctx,
				pgx.Identifier{"tf_idf_counts_stage"},
				[]string{"url_id", "token", "count"},
				pgx.CopyFromRows(countRows),
			)
			if err != nil {
				return err
			}

			// 3) Merge stage -> real tokens
			_, err = tx.Exec(ctx, `
				insert into tf_idf_tokens(token)
				select distinct s.token
				from tf_idf_tokens_stage s
				on conflict (token) do nothing;
			`)
			if err != nil {
				return err
			}

			// 4) Merge stage -> real counts using token_id lookup
			_, err = tx.Exec(ctx, `
				insert into tf_idf_counts(url_id, token_id, count)
				select
					s.url_id,
					t.id as token_id,
					s."count"
				from tf_idf_counts_stage s
				join tf_idf_tokens t
					on t.token = s.token
				on conflict (url_id, token_id) do update
				set count = excluded.count;
			`)
			if err != nil {
				return err
			}

			// 5) Clear staging
			_, err = tx.Exec(ctx, `truncate tf_idf_tokens_stage, tf_idf_counts_stage;`)
			if err != nil {
				return err
			}

			return tx.Commit(ctx)
		}()

		if err != nil {
			// Check if it's a deadlock error (SQLSTATE 40P01)
			if strings.Contains(err.Error(), "40P01") {
				time.Sleep(time.Duration(i*100) * time.Millisecond)
				continue
			}
			fmt.Fprintf(os.Stderr, "Failed to index tokens: %v\n", err)
			return
		}
		return // Success!
	}
	fmt.Fprintf(os.Stderr, "Failed to index tokens after retries\n")
}
