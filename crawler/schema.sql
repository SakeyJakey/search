create extension vector;

create table urls (
	id bigserial primary key,
	url text not null unique,
	title text not null
);

-- create table if not exists tf_idf_tokens_stage (
-- 	token text not null
-- );
--
-- create table if not exists tf_idf_counts_stage (
-- 	url_id bigint not null,
-- 	token text not null,
-- 	"count" int not null
-- );

create table tf_idf_tokens (
	id bigserial primary key,
	token text not null unique
);

create table tf_idf_counts (
	url_id bigint not null references urls(id) on delete cascade,
	token_id bigint not null references tf_idf_tokens(id) on delete cascade,
	"count" int not null check ("count" >= 0),
	primary key (url_id, token_id)
);

create table tf_idf_index (
	url_id bigint not null references urls(id) on delete cascade,
	token_id bigint not null references tf_idf_tokens(id) on delete cascade,
	tf_idf double precision not null check (tf_idf >= 0),
	primary key (url_id, token_id)
);

create table vector_index (
	url_id bigint not null references urls(id) on delete cascade,
	embedding vector(300)
);
