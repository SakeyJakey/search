package main

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
	"crawler/db"

	"golang.org/x/net/html"
)

const DATABASE_URL = "postgres://localhost/search"
const EMBEDDINGS_MAX = 300

type Statistics struct {
	mu sync.Mutex
	total int
	previousTotal int
	maxPerSecond int
	dropped int
}

var done = make(map[string]bool)
var doneMu sync.Mutex
var stats Statistics = Statistics{}
var conn *pgx.Conn
var queries *db.Queries

func getLinks(doc *html.Node) []string {
	links := make([]string, 0)

	for node := doc.FirstChild; node != nil; node = node.NextSibling {
		if node.Type == html.ElementNode {
			if node.Data == "a" {
				for _, attr := range node.Attr {
					if attr.Key == "href" {
						if strings.HasPrefix(attr.Val, "https://") {
							links = append(links, attr.Val)
						}
					}
				}
			}

			links = append(links, getLinks(node)...)
		}
	}

	return links
}

func getContent(doc *html.Node) string {
	var text strings.Builder
	for node := doc.FirstChild; node != nil; node = node.NextSibling {
		if node.Type == html.TextNode {
			text.WriteString(node.Data)
		} else {
			text.WriteString(getContent(node))
		}
	}

	return text.String()
}

func crawl(seed string, queue *Queue[string], wg *sync.WaitGroup) {
	defer wg.Done()
	doneMu.Lock()

	if done[seed] {
		doneMu.Unlock()
		return
	}

	done[seed] = true
	doneMu.Unlock()

	resp, err := http.Get(seed)

	if err != nil {
		stats.mu.Lock()
		stats.dropped += 1
		stats.mu.Unlock()
		return
	}

	defer resp.Body.Close()

	if resp.ContentLength > 100000 || !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		return
	}

	doc, err := html.Parse(resp.Body)

	if err != nil {
		fmt.Printf("Failed to parse %s\n", seed)
		return
	}

	stats.mu.Lock()
	stats.total += 1
	stats.mu.Unlock()

	links := getLinks(doc)
	content := getContent(doc)

	ctx := context.Background()
	urlID, err := queries.AddURL(ctx, seed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add URL %s to database: %v", seed, err)
	}

	embeddings_add(ctx, urlID, content)
	tfidf_add_token_index(ctx, urlID, content)

	for _, link := range links {
		wg.Add(1)
		queue.Enqueue(link)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: go run crawler <seeds file>\n")
		os.Exit(1)
	}

	seedsFile, err := os.ReadFile(os.Args[1])

	if err != nil {
		fmt.Fprintf(os.Stderr, "Seeds file invalid.\n")
		os.Exit(1)
	}

	// conn, err := pgx.Connect(context.Background(), DATABASE_URL)
	conn, err := pgxpool.New(context.Background(), DATABASE_URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	queries = db.New(conn)

	embeddings_init()

	queue := &Queue[string] { }

	var wg sync.WaitGroup

	seeds := strings.Split(string(seedsFile), "\n")

	for range 10000 {
		go func() {
			for {
				url, err := queue.Dequeue()

				if err != nil {
					break
				}

				crawl(url, queue, &wg)
			}
		}()
	}

	for _, seed := range seeds {
		if seed == "" {
			continue
		}

		wg.Add(1)
		queue.Enqueue(seed)
	}

	start := time.Now()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			elapsed := time.Since(start).String()

			stats.mu.Lock()

			percDropped := float64(stats.dropped) / math.Max(float64(stats.total+stats.dropped), 1) * 100
			totalPerSec := stats.total - stats.previousTotal

			if totalPerSec > stats.maxPerSecond {
				stats.maxPerSecond = totalPerSec
			}

			stats.previousTotal = stats.total

			fmt.Printf("Speed: %4d/sec Max Speed: %4d/sec Total: %10d Dropped: %4d (%3d%%) Elapsed: %s Queue: %10d \033[0K\r",
			            totalPerSec, stats.maxPerSecond, stats.total, stats.dropped, int(percDropped), elapsed, queue.Length())

			stats.mu.Unlock()
		}
	}()

	wg.Wait()
	fmt.Println("Done")
}
