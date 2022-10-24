package urlstorage

import (
	"context"
	"fmt"
	"sync"
)

type RTStorage struct {
	cfg Config
	ctx context.Context

	mtx sync.RWMutex
	mp  map[string]string
}

func NewRuntimeStorage(cfg Config) *RTStorage {
	return &RTStorage{
		mtx: sync.RWMutex{},
		ctx: context.Background(),
		mp:  make(map[string]string),
		cfg: cfg,
	}
}

func (s *RTStorage) Shorten(_ context.Context, url string) (string, error) {
	short := NewShortURL(url)
	var key string
	counter := 0
	for key = short.Next(s.cfg.UrlSize); s.existsURL(key); counter++ {
		if counter > s.cfg.MaxKeyGenAttempts {
			return "", fmt.Errorf("key generation attempts count exceeded %d", s.cfg.MaxKeyGenAttempts)
		}
	}
	s.mtx.Lock()
	s.mp[key] = url
	s.mtx.Unlock()
	return key, nil
}

func (s *RTStorage) Expand(_ context.Context, key string) (string, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	val, ok := s.mp[key]
	fmt.Println(s.mp)
	if !ok {
		return "", nil
	}
	return val, nil
}

func (s *RTStorage) existsURL(key string) bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	_, exists := s.mp[key]
	return exists
}
