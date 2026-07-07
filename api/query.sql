-- Vector similarity search
-- name: VectorSearch :many
SELECT
	u.id,
	u.url,
	u.title
FROM vector_index vi
JOIN urls u ON u.id = vi.url_id
WHERE vi.embedding IS NOT NULL
ORDER BY vi.embedding <#> sqlc.arg('query_embedding')::vector
LIMIT sqlc.arg('limit')::int;

-- TF-IDF search with token matching
-- name: TFIDFSearch :many
SELECT
	u.id,
	u.url,
	u.title,
	agg.relevance_score,
	agg.matched_tokens
FROM (
    SELECT url_id, SUM(tf_idf)::double precision AS relevance_score, COUNT(DISTINCT token_id) AS matched_tokens
    FROM tf_idf_index
    WHERE token_id IN (SELECT id FROM tf_idf_tokens WHERE token = ANY(sqlc.arg('tokens')::text[]))
    GROUP BY url_id
    ORDER BY relevance_score DESC, matched_tokens DESC
    LIMIT sqlc.arg('limit')::int
) AS agg
JOIN urls u ON u.id = agg.url_id;

-- name: GetTokenIds :one
SELECT 
	id,
	token
FROM tf_idf_tokens
WHERE token = $1;

-- name: SaveEvaluation :exec
INSERT INTO evaluation_results (query, category, url, search_type, is_relevant)
VALUES ($1, $2, $3, $4, $5);

-- name: GetEvaluationSummary :many
SELECT search_type, 
       category,
       COUNT(*) AS total, 
       SUM(CASE WHEN is_relevant THEN 1 ELSE 0 END) AS relevant
FROM evaluation_results
GROUP BY search_type, category;
