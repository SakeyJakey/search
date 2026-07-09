all:
	@cd crawler && go build .
	@echo [1/4] Built Crawler
	@cd postprocessing && go build .
	@echo [2/4] Built Post Processing
	@cd api && go build .
	@echo [3/4] Built API
	@cd ui && pnpm i
	@echo [4/4] Got UI Dependencies

get-embeddings:
	scripts/get-embeddings.sh

init-db:
	psql postgres -c "create database search;"
	psql search -f crawler/schema.sql
