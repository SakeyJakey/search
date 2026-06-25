package main

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"crawler/db"

	"github.com/jackc/pgx/v5/pgxpool"

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
var conn *pgxpool.Pool
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

func getTitle(doc *html.Node) string {
	for node := doc.FirstChild; node != nil; node = node.NextSibling {
		if node.Type == html.ElementNode {
			if node.Data == "title" {
				for child := node.FirstChild; child != nil; child = child.NextSibling {
					if child.Type == html.TextNode {
						return child.Data
					}
				}
			} else {




				title := getTitle(node)
				if title != "" {
					return title 
				}
			}
		}
	}

	return "Untitled Page"
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

	req, err := http.NewRequest("GET", seed, nil)
	if err != nil {
		stats.mu.Lock()
		stats.dropped += 1
		stats.mu.Unlock()
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.3")

	resp, err := http.DefaultClient.Do(req)

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
	title := getTitle(doc)

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	urlID, err := queries.AddURL(ctx, db.AddURLParams {
		Url: seed,
		Title: title,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add URL %s to database: %v", seed, err)
	}

	err = queries.AddRawContent(ctx, db.AddRawContentParams{
		UrlID: urlID,
		Content: content,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add raw content for URL %s to database: %v", seed, err)
	}

	for _, link := range links {
		wg.Add(1)
		queue.Enqueue(link)
	}

	if err := ctx.Err(); err != nil {
		queue.Enqueue(seed)
		return
	}

	wg.Done()
}

func main() {
	seedsPath := "seeds.txt"

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "No seeds file found\n")
		fmt.Printf("Defaulting to seeds.txt")
	} else {
		seedsPath = os.Args[1]
	}

	seedsFile, err := os.ReadFile(seedsPath)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Seeds file invalid.\n")
		os.Exit(1)
	}

	conn, err = pgxpool.New(context.Background(), DATABASE_URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Clearing database...")
	_, err = conn.Exec(context.Background(), "TRUNCATE TABLE urls RESTART IDENTITY CASCADE;")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to clear database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Database cleared.")

	queries = db.New(conn)

	queue := &Queue[string] { }

	var wg sync.WaitGroup

	seeds := strings.Split(string(seedsFile), "\n")

	for range runtime.NumCPU() {
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
