package my_redis

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v7"
)

type RedisClient struct {
	pool *sync.Pool
}

func NewRedisClient(addr string) *RedisClient {
	return &RedisClient{
		pool: &sync.Pool{
			New: func() interface{} {
				return redis.NewClient(&redis.Options{
					Addr:     addr,
					Password: "",
					DB:       0,
				})
			},
		},
	}
}

func (rc *RedisClient) XRead(streamKeys []string, ids []string, count int, block time.Duration) ([]redis.XStream, error) {
	client := rc.pool.Get().(*redis.Client)
	defer rc.pool.Put(client)
	return client.XRead(&redis.XReadArgs{
		Streams: append(streamKeys, ids...),
		Count:   int64(count),
		Block:   block,
	}).Result()
}

func (rc *RedisClient) XReadOne(streamKeys []string, ids []string, count int, block time.Duration) (map[string]interface{}, string, error) {
	client := rc.pool.Get().(*redis.Client)
	defer rc.pool.Put(client)
	ret, err := client.XRead(&redis.XReadArgs{
		Streams: append(streamKeys, ids...),
		Count:   int64(count),
		Block:   block,
	}).Result()
	if err != nil {
		return nil, "", fmt.Errorf("cant read by stream: %v", err)
	}
	if len(ret) == 0 {
		return nil, "", fmt.Errorf("stream length is zero: %v", err)
	}
	if len(ret[0].Messages) == 0 {
		return nil, "", fmt.Errorf("message length is zero: %v", err)
	}
	return ret[0].Messages[0].Values, ret[0].Messages[0].ID, nil
}

func (rc *RedisClient) XAdd(streamKey string, value map[string]interface{}) error {
	client := rc.pool.Get().(*redis.Client)
	defer rc.pool.Put(client)
	_, err := client.XAdd(&redis.XAddArgs{
		Stream: streamKey,
		MaxLen: 50,
		Values: value,
	}).Result()

	if err != nil {
		return err
	}

	return nil
}

func (rc *RedisClient) FlushAll() error {
	client := rc.pool.Get().(*redis.Client)
	defer rc.pool.Put(client)
	_, err := client.FlushAll().Result()
	return err
}

func (rc *RedisClient) XDel(streamKey string, ids []string) error {
	client := rc.pool.Get().(*redis.Client)
	defer rc.pool.Put(client)
	_, err := client.XDel(streamKey, ids...).Result()
	return err
}

func (rc *RedisClient) HGet(key string) (map[string]string, error) {
	client := rc.pool.Get().(*redis.Client)
	defer rc.pool.Put(client)
	return client.HGetAll(key).Result()
}

func (rc *RedisClient) HSet(key string, value map[string]interface{}) (int64, error) {
	client := rc.pool.Get().(*redis.Client)
	defer rc.pool.Put(client)
	count, err := client.HSet(key, value).Result()
	return count, err
}

func (rc *RedisClient) Delete(keys ...string) (int64, error) {
	client := rc.pool.Get().(*redis.Client)
	defer rc.pool.Put(client)
	return client.Del(keys...).Result()
}
