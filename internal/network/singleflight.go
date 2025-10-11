package network

import (
	"context"
	"sync"
)

// SingleFlight предотвращает дублирование одинаковых запросов
type SingleFlight struct {
	mu sync.Mutex
	m  map[string]*call
}

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// NewSingleFlight создает новый SingleFlight
func NewSingleFlight() *SingleFlight {
	return &SingleFlight{
		m: make(map[string]*call),
	}
}

// Do выполняет функцию только один раз для одинаковых ключей
func (sf *SingleFlight) Do(ctx context.Context, key string, fn func() (interface{}, error)) (interface{}, error) {
	sf.mu.Lock()
	if c, ok := sf.m[key]; ok {
		sf.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := new(call)
	c.wg.Add(1)
	sf.m[key] = c
	sf.mu.Unlock()

	// Выполняем функцию
	c.val, c.err = fn()
	c.wg.Done()

	// Удаляем из карты после завершения
	sf.mu.Lock()
	delete(sf.m, key)
	sf.mu.Unlock()

	return c.val, c.err
}
