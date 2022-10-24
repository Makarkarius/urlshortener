package urlstorage

import (
	"context"
	"fmt"
)

// TODO: add logger

type StorageType bool

const (
	RuntimeStorage  StorageType = false
	DatabaseStorage StorageType = true
)

type Config struct {
	StorageType       StorageType
	UrlSize           int
	MaxKeyGenAttempts int
	ShortenTimeoutSec int
	ExpandTimeoutSec  int

	DBCfg *DBStorageConfig
}

type Storage interface {
	Shorten(ctx context.Context, url string) (string, error)
	Expand(ctx context.Context, key string) (string, error)
}

func NewStorage(cfg Config) (Storage, error) {
	if err := cfg.verify(); err != nil {
		return nil, fmt.Errorf("invalid cfg: %w", err)
	}
	if cfg.StorageType == DatabaseStorage {
		return NewDBStorage(cfg)
	}
	return NewRuntimeStorage(cfg), nil
}

func (c Config) verify() error {
	if c.UrlSize < 2 {
		return fmt.Errorf("url size can't be smaller than 2")
	}
	if c.StorageType == DatabaseStorage && c.DBCfg == nil {
		return fmt.Errorf("database config is nil")
	}
	return nil
}
