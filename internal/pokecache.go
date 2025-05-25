package internal

import (
	"fmt"
	"sync"
	"time"
)

type CacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	Mute     sync.Mutex
	CacheMap map[string]CacheEntry
}

func NewCache(interval time.Duration) *Cache {
	fmt.Printf("interval : %v\n", interval)
	newCache := Cache{Mute: sync.Mutex{}, CacheMap: make(map[string]CacheEntry, 0)}
	go newCache.reapLoop(interval)
	return &newCache
}

func (p *Cache) Add(name string, val []byte) {
	p.Mute.Lock()
	defer p.Mute.Unlock()
	p.CacheMap[name] = CacheEntry{createdAt: time.Now(), val: val}
}

func (p *Cache) Get(name string) ([]byte, bool) {
	p.Mute.Lock()
	defer p.Mute.Unlock()
	entry, ok := p.CacheMap[name]
	if !ok {
		return nil, false
	}
	return entry.val, true
}

func (p *Cache) reapLoop(interval time.Duration) {
	p.Mute.Lock()
	defer p.Mute.Unlock()

	for name, _ := range p.CacheMap {
		if time.Since(p.CacheMap[name].createdAt) > interval {
			delete(p.CacheMap, name)
		}

	}
}
