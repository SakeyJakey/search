
-- below was created with AI

-- Vector index (IVFFlat for cosine similarity search)
CREATE INDEX ON vector_index USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- TF-IDF indices
CREATE INDEX idx_tfidf_token_id ON tf_idf_index(token_id);
CREATE INDEX idx_tfidf_url_id ON tf_idf_index(url_id);
CREATE INDEX idx_tfidf_token_score ON tf_idf_index(token_id, tf_idf DESC);

-- Token lookup
CREATE INDEX idx_token_text ON tf_idf_tokens(token);

