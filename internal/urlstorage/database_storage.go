package urlstorage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"time"
)

// TODO: implement with sharding
// TODO: implement with caching

type DBStorageConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type DBStorage struct {
	cfg Config
	ctx context.Context

	db *pgx.Conn
}

func NewDBStorage(cfg Config) (*DBStorage, error) {
	conn, err := pgx.Connect(context.Background(),
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.DBCfg.Username, cfg.DBCfg.Password, cfg.DBCfg.Host, cfg.DBCfg.Port, cfg.DBCfg.Database))
	if err != nil {
		return nil, err
	}
	return &DBStorage{
		cfg: cfg,
		ctx: context.Background(),
		db:  conn}, nil
}

func (s *DBStorage) Shorten(ctx context.Context, url string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.ShortenTimeoutSec)*time.Second)
	defer cancel()

	short := NewShortURL(url)
	var key string
	counter := 0
	for ; ; counter++ {
		key = short.Next(s.cfg.UrlSize)
		ok, err := s.existsURL(ctx, key)
		if err != nil {
			return "", err
		}
		if !ok {
			break
		}
		if counter > s.cfg.MaxKeyGenAttempts {
			return "", fmt.Errorf("key generation attempts count exceeded %d", s.cfg.MaxKeyGenAttempts)
		}
	}
	err := s.insertUrl(ctx, url, key)
	return key, err
}

func (s *DBStorage) Expand(ctx context.Context, key string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.ExpandTimeoutSec)*time.Second)
	defer cancel()

	val, err := s.getUrl(ctx, key)
	if err != nil {
		return "", err
	}
	return val, nil
}

func (s *DBStorage) existsURL(ctx context.Context, key string) (ok bool, _ error) {
	row := s.db.QueryRow(ctx, "SELECT EXISTS (SELECT * FROM URLS WHERE key = $1)", key)
	err := row.Scan(&ok)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (s *DBStorage) insertUrl(ctx context.Context, url, key string) error {
	_, err := s.db.Exec(ctx, "INSERT INTO URLS(url, key) VALUES ($1, $2)", url, key)
	return err
}

func (s *DBStorage) getUrl(ctx context.Context, key string) (url string, _ error) {
	row := s.db.QueryRow(ctx, "SELECT url FROM URLS WHERE key = $1", key)
	err := row.Scan(&url)
	if err != nil {
		return "", err
	}
	return url, nil
}
