package main

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type Statistics struct {
	mu sync.Mutex
	total int
	previousTotal int
	maxPerSecond int
	dropped int
}

type Queue[T any] struct {
	mu sync.Mutex
	items []T
}

func (q *Queue[T]) Enqueue(item T) {
	q.mu.Lock()
	q.items = append(q.items, item)
	q.mu.Unlock()
}

func (q *Queue[T]) Dequeue() (T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var null T
	if len(q.items) == 0 {
		return null, errors.New("Queue empty")
	}

	item := q.items[0]
	q.items = q.items[1:]

	return item, nil
}

func (q *Queue[T]) Length() int {
	q.mu.Lock()
	length := len(q.items)
	q.mu.Unlock()
	return length
}

var done = make(map[string]bool)
var doneMu sync.Mutex

var stats Statistics = Statistics{}

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

func crawl(seed string, queue *Queue[string], wg *sync.WaitGroup, output *os.File) {
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

	fmt.Fprintf(output, "%s\n", seed)
	stats.mu.Lock()
	stats.total += 1
	stats.mu.Unlock()

	links := getLinks(doc)

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


	output, err := os.Create("output.txt")

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output file.\n")
	}

	defer output.Close()

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

				crawl(url, queue, &wg, output)
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
