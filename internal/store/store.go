package store

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type SessionStore struct {
	cache *cache.Cache
}

func New(ttl time.Duration) *SessionStore {
	return &SessionStore{
		cache: cache.New(ttl, ttl/2),
	}
}

func (s *SessionStore) Get(id string) (interface{}, bool) {
	return s.cache.Get(id)
}

func (s *SessionStore) Set(id string, data interface{}) {
	s.cache.Set(id, data, cache.DefaultExpiration)
}
