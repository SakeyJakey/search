-- name: GetDocumentCount :one
select count(distinct url_id) as total
from tf_idf_counts;

-- name: GetTokenIDs :many
select id
from tf_idf_tokens;

-- name: GetDocumentFrequencyByToken :one
select count(distinct url_id) as df
from tf_idf_counts
where token_id = $1;

-- name: GetCountsForToken :many
select url_id, "count"
from tf_idf_counts
where token_id = $1;

-- name: InsertTfIdfBatch :exec
insert into tf_idf_index (
	url_id, token_id, tf_idf
) values (
	$1, $2, $3
);
