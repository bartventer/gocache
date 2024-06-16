package ramcache

import (
	"slices"
	"sync"
	"time"
)

// Item is a cache Item.
type Item struct {
	Value  []byte    // Value is the item value.
	Expiry time.Time // Expiry is the item expiry time. Default is 24 hours.
}

// IsExpired returns true if the item is expired.
func (i Item) IsExpired() bool {
	return time.Now().After(i.Expiry)
}

// store is an in-memory store for cache items.
type store struct {
	mu    sync.RWMutex
	items map[string]Item
}

// newStore creates a new store.
func newStore() *store {
	return &store{
		items: make(map[string]Item),
	}
}

func (s *store) Get(key string) (Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, exists := s.items[key]
	return item, exists
}

func (s *store) Set(key string, item Item) {
	s.mu.Lock()
	s.items[key] = item
	s.mu.Unlock()
}

func (s *store) Delete(key string) {
	s.mu.Lock()
	delete(s.items, key)
	s.mu.Unlock()
}

func (s *store) Clear() {
	s.mu.Lock()
	s.items = make(map[string]Item)
	s.mu.Unlock()
}

// keyItem is a struct that contains a key and an item.
type keyItem struct {
	Key  string // Key is the item key.
	Item Item   // Item is the item.
}

// KeyItemsSortedByExpiry returns all key items sorted by expiry time (closest to expiry first).
func (s *store) KeyItemsSortedByExpiry() []keyItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]keyItem, 0, len(s.items))
	for key, item := range s.items {
		items = append(items, keyItem{Key: key, Item: item})
	}
	slices.SortFunc(items, func(a, b keyItem) int {
		return a.Item.Expiry.Compare(b.Item.Expiry)
	})
	return items
}