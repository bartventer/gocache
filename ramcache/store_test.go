package ramcache

import (
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	t.Run("Set", func(t *testing.T) {
		s := newStore()
		s.Set("key1", item{Value: []byte("value1"), Expiry: time.Now().Add(10 * time.Minute)})
		item, exists := s.Get("key1")
		if !exists || string(item.Value) != "value1" {
			t.Errorf("Set failed. Expected value1, got %v", string(item.Value))
		}
	})

	t.Run("Get", func(t *testing.T) {
		s := newStore()
		s.Set("key1", item{Value: []byte("value1"), Expiry: time.Now().Add(10 * time.Minute)})
		item, exists := s.Get("key1")
		if !exists || string(item.Value) != "value1" {
			t.Errorf("Get failed. Expected value1, got %v", string(item.Value))
		}
	})

	t.Run("Delete", func(t *testing.T) {
		s := newStore()
		s.Set("key1", item{Value: []byte("value1"), Expiry: time.Now().Add(10 * time.Minute)})
		s.Delete("key1")
		_, exists := s.Get("key1")
		if exists {
			t.Errorf("Delete failed. Expected key1 to be deleted")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		s := newStore()
		s.Set("key1", item{Value: []byte("value1"), Expiry: time.Now().Add(10 * time.Minute)})
		s.Set("key2", item{Value: []byte("value2"), Expiry: time.Now().Add(20 * time.Minute)})
		s.Clear()
		_, exists1 := s.Get("key1")
		_, exists2 := s.Get("key2")
		if exists1 || exists2 {
			t.Errorf("Clear failed. Expected all keys to be deleted")
		}
	})

	t.Run("KeyItemsSortedByExpiry", func(t *testing.T) {
		s := newStore()
		s.Set("key1", item{Value: []byte("value1"), Expiry: time.Now().Add(20 * time.Minute)})
		s.Set("key2", item{Value: []byte("value2"), Expiry: time.Now().Add(10 * time.Minute)})
		s.Set("key3", item{Value: []byte("value3"), NoExpiry: true})
		items := s.KeyItemsSortedByExpiry()
		if len(items) != 3 || items[0].Key != "key2" || items[1].Key != "key1" || items[2].Key != "key3" {
			t.Errorf("KeyItemsSortedByExpiry failed. Expected [key2, key1, key3], got [%v, %v, %v]", items[0].Key, items[1].Key, items[2].Key)
		}
	})
}
