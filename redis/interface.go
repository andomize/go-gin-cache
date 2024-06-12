package redis

import (
	"context"
	"time"
)

// Pool maintains a pool of Redis connections.
type Pool interface {
	Get(ctx context.Context) Conn
}

// Conn is a single Redis connection.
type Conn interface {
	Get(key string) (*Content, bool, error)
	Set(key string, content *Content, expiration time.Duration) error
	Del(key string) error
	Close() error
}

type Content struct {
	StatusCode  int    `json:"status_code"`
	ContentType string `json:"content_type"`
	Data        []byte `json:"data"`
}
