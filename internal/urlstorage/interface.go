package urlstorage

import (
	"context"
	"fmt"
)

type StorageType uint8

const (
	RuntimeStorage  StorageType = 0
	DatabaseStorage StorageType = 1
)

type Config struct {
	StorageType       StorageType `json:"storageType"`
	UrlSize           int         `json:"urlSize"`
	MaxKeyGenAttempts int         `json:"maxKeyGenAttempts"`
	ShortenTimeoutSec int         `json:"shortenTimeoutSec"`
	ExpandTimeoutSec  int         `json:"expandTimeoutSec"`

	DBCfg *DBStorageConfig `json:"DBCfg"`
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
