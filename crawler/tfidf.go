package main

import (
	"context"
	"crawler/db"
	"fmt"
	"os"
	"strings"
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
	for token, count := range index {
		tokenID, err := queries.AddToken(ctx, token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add token %s to database: %v\n", token, err)
			continue
		}

		_, err = queries.AddCount(ctx, db.AddCountParams {
			UrlID: urlID,
			TokenID: tokenID,
			Count: count,
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add count of token %s to database: %v\n", token, err)
		}
	}

}
