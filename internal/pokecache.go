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
	ticker := time.NewTicker(interval)
	for {
		if len(p.CacheMap) > 0 {

			select {
			case tick := <-ticker.C:
				p.Mute.Lock()

				for name := range p.CacheMap {
					after := tick.After(p.CacheMap[name].createdAt)
					if after {

						delete(p.CacheMap, name)
					}
				}
				p.Mute.Unlock()
			default:
				continue
			}
		}
	}
}
