all:
	%echo "Scripts:"
	%echo "- get-embeddings: download GloVe embeddings and convert them into the correct format"
	%echo "- init-db: creates the tables from the schema"

get-embeddings:
	scripts/get-embeddings.sh

init-db:
	psql -d search -f crawler/schema.sql

