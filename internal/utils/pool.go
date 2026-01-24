package utils

import (
	"sync"
)

type Resetter interface {
	Reset()
}

type Pool[T Resetter] struct {
	mu    sync.Mutex
	items []T
}

func New[T Resetter]() *Pool[T] {
	return &Pool[T]{
		items: make([]T, 0),
	}
}

func (p *Pool[T]) Get() T {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.items) == 0 {
		var empty T
		return empty
	}

	item := p.items[len(p.items)-1]
	p.items = p.items[:len(p.items)-1]
	return item
}

func (p *Pool[T]) Put(item T) {
	item.Reset()

	p.mu.Lock()
	defer p.mu.Unlock()
	p.items = append(p.items, item)
}
