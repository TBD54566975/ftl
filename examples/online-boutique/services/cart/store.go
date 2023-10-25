package cart

import (
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
)

type Store struct {
	lock  sync.Mutex
	carts *lru.Cache[string, []Item]
}

func NewStore() *Store {
	cache, err := lru.New[string, []Item](100)
	if err != nil {
		panic(err)
	}
	return &Store{carts: cache}
}

func (s *Store) Add(userID string, newItem Item) {
	s.lock.Lock()
	defer s.lock.Unlock()
	items, ok := s.carts.Get(userID)
	if !ok {
		s.carts.Add(userID, []Item{newItem})
		return
	}

	found := false
	for i, existingItem := range items {
		if existingItem.ProductID == newItem.ProductID {
			items[i].Quantity += newItem.Quantity
			found = true
			break
		}
	}

	if !found {
		items = append(items, newItem)
	}

	s.carts.Add(userID, items)
}

func (s *Store) Get(userID string) []Item {
	s.lock.Lock()
	defer s.lock.Unlock()
	items, ok := s.carts.Get(userID)
	if !ok {
		return []Item{}
	}
	out := make([]Item, len(items))
	copy(out, items)
	return out
}

func (s *Store) Empty(userID string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.carts.Remove(userID)
}
