-- name: AddVector :one
insert into vector_index (
	url_id, embedding
) values (
	$1, $2
)
returning *;

-- name: AddURL :one
insert into urls (
	url, title
) values (
	$1, $2
) on conflict (
	url
) do update set url = urls.url
returning id;

-- name: AddToken :one
insert into tf_idf_tokens (
	token
) values (
	$1
) on conflict (
	token
) do update set token = tf_idf_tokens.token
returning id;

-- name: AddCount :one
insert into tf_idf_counts (
	url_id, token_id, "count"
) values (
	$1, $2, $3
) on conflict (
	url_id, token_id
) do update set "count" = excluded."count"
returning *;

-- name: AddRawContent :exec
insert into raw_content (
	url_id, content
) values (
	$1, $2
);
