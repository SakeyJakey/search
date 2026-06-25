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
	SUM(ti.tf_idf)::double precision AS relevance_score,
	COUNT(DISTINCT ti.token_id) AS matched_tokens
FROM tf_idf_index ti
JOIN urls u ON u.id = ti.url_id
JOIN tf_idf_tokens t ON t.id = ti.token_id
WHERE t.token = ANY(sqlc.arg('tokens')::text[])
GROUP BY u.id, u.url, u.title
ORDER BY relevance_score DESC, matched_tokens DESC
LIMIT sqlc.arg('limit')::int;

-- name: GetTokenIds :one
SELECT 
	id,
	token
FROM tf_idf_tokens
WHERE token = $1;

-- name: SaveEvaluation :exec
INSERT INTO evaluation_results (query, url, search_type, is_relevant)
VALUES ($1, $2, $3, $4);

-- name: GetEvaluationSummary :many
SELECT search_type, 
       COUNT(*) AS total, 
       SUM(CASE WHEN is_relevant THEN 1 ELSE 0 END) AS relevant
FROM evaluation_results
GROUP BY search_type;
