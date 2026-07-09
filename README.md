# Assessing the Relevance of Search Results Retrieved Through TF-IDF Compared to Meaning-Based Vector Embeddings

## Requirements
- Go
- pnpm
- PostgreSQL
- POSIX environment
- 200GB of storage per million sites

## How to Run

- `make` to compile all binaries and get dependencies
- `make init-db` to initialise the database
- `make get-embeddings` to download the pre-trained GloVe embeddings
- `./crawler/crawler` to begin discovering sites
- Once the crawler has reached the desired number of sites, you can run `./postprocessing/postprocessing` to create the index
- After this completes, run `./api/api` and `cd ui/artifacts/search-ui/ && pnpm run dev` to run the search engine UI
- Then, once the API has finished loading the embeddings, the search engine is ready for testing
