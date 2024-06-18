package ramcache

import (
	"slices"
	"sync"
	"time"
)

// item is a cache item.
type item struct {
	Value    []byte    // Value is the item value.
	Expiry   time.Time // Expiry is the item expiry time. Default is 24 hours.
	NoExpiry bool      // NoExpiry indicates whether the item should never expire.
}

// IsExpired returns true if the item is expired.
func (i item) IsExpired() bool {
	if i.NoExpiry {
		return false
	}
	return time.Now().After(i.Expiry)
}

// store is an in-memory store for cache items.
type store struct {
	mu    sync.RWMutex
	items map[string]item
}

// newStore creates a new store.
func newStore() *store {
	return &store{
		items: make(map[string]item),
	}
}

func (s *store) Get(key string) (item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, exists := s.items[key]
	return item, exists
}

func (s *store) Set(key string, item item) {
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
	s.items = make(map[string]item)
	s.mu.Unlock()
}

// keyItem is a struct that contains a key and an item.
type keyItem struct {
	Key  string // Key is the item key.
	Item item   // Item is the item.
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
		if a.Item.NoExpiry && b.Item.NoExpiry {
			return 0
		}
		if a.Item.NoExpiry {
			return 1
		}
		if b.Item.NoExpiry {
			return -1
		}
		return a.Item.Expiry.Compare(b.Item.Expiry)
	})
	return items
}
