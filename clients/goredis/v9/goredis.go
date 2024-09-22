package goredis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	gincacheclients "github.com/andomize/go-gin-cache/clients"
	"github.com/redis/go-redis/v9"
)

type pool struct {
	delegate redis.UniversalClient
}

func (p *pool) Get(ctx context.Context) gincacheclients.Conn {
	if ctx == nil {
		ctx = context.Background()
	}
	return &conn{p.delegate, ctx}
}

// NewPool returns a Goredis-based pool implementation.
func NewPool(delegate redis.UniversalClient) gincacheclients.Pool {
	return &pool{delegate}
}

type conn struct {
	delegate redis.UniversalClient
	ctx      context.Context
}

func (c *conn) Get(key string) (*gincacheclients.Content, bool, error) {
	value, err := c.delegate.Get(c.ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	content := gincacheclients.Content{}
	err = json.Unmarshal([]byte(value), &content)
	if err != nil {
		return nil, false, err
	}
	return &content, true, nil
}

func (c *conn) Set(key string, content *gincacheclients.Content, expiration time.Duration) error {
	bytes, _ := json.Marshal(content)
	_, err := c.delegate.Set(c.ctx, key, bytes, 0).Result()
	return err
}

func (c *conn) Del(prefix string) error {
	keys, err := c.delegate.Keys(c.ctx, fmt.Sprint(prefix, "*")).Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}

	err = c.delegate.Del(c.ctx, keys...).Err()
	// if err == redis.Nil {
	// 	return nil
	// }
	if err != nil {
		return err
	}

	return err
}

func (c *conn) Close() error {
	return c.delegate.Close()
}
