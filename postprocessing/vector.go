package main

import (
	"bufio"
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

var embeddings = make(map[string][]float32)

func vecMag(a []float32) float64 {
	sum := 0.0
	for _, val := range a {
		sum += float64(val * val)
	}
	return math.Sqrt(sum)
}

func vecDot(a []float32, b []float32) float64 {
	sum := 0.0
	for i := range a {
		sum += float64(a[i] * b[i])
	}
	return sum
}

func vecSim(a []float32, b []float32) float64 {
	return vecDot(a, b) / (vecMag(a) * vecMag(b))
}

func vecAdd(a []float32, b []float32) []float32 {
	for i, val := range b {
		a[i] += val
	}
	return a
}

func vecSum(vec []float32) float32 {
	sum := float32(0.0)
	for _, val := range vec {
		sum += val
	}
	return sum
}

func vecDiv(a []float32, b []float32) []float32 {
	for i, val := range b {
		a[i] /= val
	}
	return a
}

func vecFill(val float32, size int) []float32 {
	vec := make([]float32, size)
	for i := range vec {
		vec[i] = val
	}
	return vec
}

func vecNorm(vec []float32) []float32 {
	mag := float32(vecMag(vec))
	if mag == 0 {
		return vec
	}
	for i := range vec {
		vec[i] /= mag
	}
	return vec
}

func vec32(orig []float32) []float32 {
    return orig
}

func embeddings_init() {
	start := time.Now()
	fmt.Println("Loading embeddings...")

	file, err := os.Open("../data/dolma_300_2024_1.2M.100_combined.txt")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to open embeddings file");
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		list := strings.Split(line, " ")
		if len(list) < 2 {
			continue
		}
		
		vector := make([]float32, len(list) - 1)
		
		for i, num := range list[1:] {
			val, err := strconv.ParseFloat(num, 32)
			if err != nil {
				vector[i] = 0
				continue
			}
			vector[i] = float32(val)
		}

		embeddings[list[0]] = vector
		lineCount++
		if lineCount%10000 == 0 {
			fmt.Printf("Loaded %d lines...\n", lineCount)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading embeddings file:", err)
		return
	}

	fmt.Printf("Loaded %d embeddings\n", lineCount)
	fmt.Printf("Took %.1fseconds\n", time.Since(start).Seconds())
}

func embeddings_embed(text string) []float32 {
	embedding := make([]float32, EMBEDDINGS_MAX)
	count := 0

	for token := range strings.FieldsSeq(text) {
		
		embedding = vecAdd(embedding, embeddings[token])
		count++
	}

	embedding = vecDiv(embedding, vecFill(float32(count), EMBEDDINGS_MAX))

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
