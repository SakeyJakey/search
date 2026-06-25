package main

import (
	"context"
	"postprocessing/db"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pgvector/pgvector-go"
)

var embeddings = make(map[string][]float64)

func vecMag(a []float64) float64 {
	sum := 0.0
	for _, val := range a {
		sum += val * val
	}
	return math.Sqrt(sum)
}

func vecDot(a []float64, b []float64) float64 {
	sum := 0.0
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

func vecSim(a []float64, b []float64) float64 {
	return vecDot(a, b) / (vecMag(a) * vecMag(b))
}

func vecAdd(a []float64, b []float64) []float64 {
	for i, val := range b {
		a[i] += val
	}
	return a
}

func vecSum(vec []float64) float64 {
	sum := 0.0
	for _, val := range vec {
		sum += val
	}
	return sum
}

func vecDiv(a []float64, b []float64) []float64 {
	for i, val := range b {
		a[i] /= val
	}
	return a
}

func vecFill(val float64, size int) []float64 {
	vec := make([]float64, size)
	for i := range vec {
		vec[i] = val
	}
	return vec
}

func vecNorm(vec []float64) []float64 {
	mag := vecMag(vec)
	if mag == 0 {
		return vec
	}
	for i := range vec {
		vec[i] /= mag
	}
	return vec
}

func vec32(orig[]float64) []float32 {
	vec := make([]float32, len(orig))
	for i, val := range orig {
		vec[i] = float32(val)
	}
	return vec
}

func embeddings_init() {
	start := time.Now()
	fmt.Println("Loading embeddings...")

	data, err := os.ReadFile("../data/dolma_300_2024_1.2M.100_combined.txt")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read embeddings");
		return
	}

	lines := strings.SplitSeq(string(data), "\n")

	for line := range lines {
		list := strings.Split(line, " ")
		vector := make([]float64, len(list) - 1)
		
		for i, num := range list[1:] {
			vector[i], err = strconv.ParseFloat(num, 64)
			if err != nil {
				vector[i] = 0
				fmt.Fprintln(os.Stderr, "Failed to parse embeddings");
				return
			}
		}

		embeddings[list[0]] = vector
	}

	fmt.Println("Loaded embeddings")

	fmt.Printf("Took %.1fseconds\n", time.Since(start).Seconds())
}

func embeddings_embed(text string) []float64 {
	embedding := make([]float64, EMBEDDINGS_MAX)
	count := 0

	for token := range strings.FieldsSeq(text) {
		
		embedding = vecAdd(embedding, embeddings[token])
		count++
	}

	embedding = vecDiv(embedding, vecFill(float64(count), EMBEDDINGS_MAX))

	return embedding
}

func embeddings_add(ctx context.Context, urlID int64, content string) {
	_, err := queries.AddVector(ctx, db.AddVectorParams{
		UrlID: urlID,
		Embedding: pgvector.NewVector(vec32(vecNorm(embeddings_embed(content)))),
	})
	
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add vector to database: %v\n", err)
	}

}
