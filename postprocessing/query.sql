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

-- name: GetUnprocessedRawContent :many
select url_id, content
from raw_content;

-- name: MoveToProcessed :exec
insert into processed_content (url_id, content)
select r.url_id, r.content from raw_content r where r.url_id = $1;

-- name: DeleteRawContent :exec
delete from raw_content where url_id = $1;

-- name: AddVector :one
insert into vector_index (
	url_id, embedding
) values (
	$1, $2
)
returning *;
