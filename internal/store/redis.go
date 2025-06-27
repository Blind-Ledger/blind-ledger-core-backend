package store

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type RedisStore struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisStore(addr, pass string, db int) *RedisStore {
	return &RedisStore{
		client: redis.NewClient(&redis.Options{Addr: addr, Password: pass, DB: db}),
		ctx:    context.Background(),
	}
}

func (r *RedisStore) Publish(msg Message) error {
	return r.client.Publish(r.ctx, msg.Channel, msg.Data).Err()
}

func (r *RedisStore) Subscribe(channel string) (<-chan Message, error) {
	sub := r.client.Subscribe(r.ctx, channel)
	ch := make(chan Message)
	go func() {
		for v := range sub.Channel() {
			ch <- Message{Channel: v.Channel, Data: []byte(v.Payload)}
		}
	}()
	return ch, nil
}
