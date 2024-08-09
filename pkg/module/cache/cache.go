package cache

import (
	"sync"
	"time"
)

type TinyCache interface {
	Set(key string, val any)
	SetExpired(key string, val any, expiredAt time.Duration)
	Get(key string) (any, bool)
	BurnAfterGet(key string) (any, bool)
	Delete(key string)
}

type memoryCache struct {
	store map[string]*memoryValue
	// expiredCh chan *memoryValue
	opCh chan func()
	mu   sync.Mutex
}

type memoryValue struct {
	key      string
	data     any
	expireAt int64
}

// 仿造redis实现的 单goroutine 内存缓存
func NewMemoryCache() TinyCache {
	mc := &memoryCache{
		store: make(map[string]*memoryValue),
		// expiredCh: make(chan *memoryValue),
		opCh: make(chan func(), 4),
		mu:   sync.Mutex{},
	}

	go mc.run()

	return mc
}

func (m *memoryCache) run() {
	// 每小时清理一次
	t := time.NewTicker(time.Hour)

	for {
		select {
		case <-t.C:
			m.opCh <- func() {
				now := time.Now().Unix()
				m.mu.Lock()
				defer m.mu.Unlock()
				for key, val := range m.store {
					if val.expireAt != 0 && val.expireAt <= now {
						delete(m.store, key)
					}
				}
			}
		case op := <-m.opCh:
			op()
		}
	}
}

func (m *memoryCache) Set(key string, value any) {
	m.opCh <- func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.store[key] = &memoryValue{key: key, data: value}
	}
}

func (m *memoryCache) SetExpired(key string, value any, expiredAt time.Duration) {
	expireAt := time.Now().Add(expiredAt).Unix()
	m.opCh <- func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.store[key] = &memoryValue{key: key, data: value, expireAt: expireAt}
	}
}

func (m *memoryCache) Get(key string) (any, bool) {
	resultCh := make(chan any)
	m.opCh <- func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		val, ok := m.store[key]
		if !ok {
			resultCh <- nil
			return
		}
		if val.expireAt != 0 && val.expireAt < time.Now().Unix() {
			// 已过期
			delete(m.store, key)
			resultCh <- nil
			return
		}
		resultCh <- val.data
	}
	result := <-resultCh
	if result == nil {
		return nil, false
	}

	return result, true
}

func (m *memoryCache) BurnAfterGet(key string) (any, bool) {
	resultCh := make(chan any)
	m.opCh <- func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		val, ok := m.store[key]
		if !ok {
			resultCh <- nil
			return
		}
		if val.expireAt != 0 && val.expireAt < time.Now().Unix() {
			// 已过期
			delete(m.store, key)
			resultCh <- nil
			return
		}
		// Burn after reading
		delete(m.store, key)
		resultCh <- val.data
	}
	result := <-resultCh
	if result == nil {
		return nil, false
	}

	return result, true
}

func (m *memoryCache) Delete(key string) {
	m.opCh <- func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		delete(m.store, key)
	}
}
