package main

import (
	"errors"
	"sync"
)

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

