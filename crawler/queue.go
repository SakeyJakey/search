package main

import (
	"sync"
)

type Queue[T any] struct {
	mu    sync.Mutex
	cond  *sync.Cond
	items []T
}

func (q *Queue[T]) Enqueue(item T) {
	q.mu.Lock()
	if q.cond == nil {
		q.cond = sync.NewCond(&q.mu)
	}
	q.items = append(q.items, item)
	q.cond.Signal()
	q.mu.Unlock()
}

func (q *Queue[T]) Dequeue() T {
	q.mu.Lock()
	if q.cond == nil {
		q.cond = sync.NewCond(&q.mu)
	}
	for len(q.items) == 0 {
		q.cond.Wait()
	}
	item := q.items[0]
	q.items = q.items[1:]
	q.mu.Unlock()
	return item
}

func (q *Queue[T]) Length() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

