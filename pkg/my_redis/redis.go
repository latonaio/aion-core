package my_redis

import (
	"fmt"
	"time"

	"bitbucket.org/latonaio/aion-core/pkg/log"
	"github.com/avast/retry-go"
	"github.com/go-redis/redis/v7"
)

const (
	connRetryCount = 10
)

type RedisClient struct {
	client *redis.Client
}

func GetInstance() *RedisClient {
	return &RedisClient{}
}

func (rc *RedisClient) CreatePool(addr string) error {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	// redis connection check
	if err := retry.Do(
		func() error {
			if _, err := client.Ping().Result(); err != nil {
				return fmt.Errorf("cant connect to redis (Address: %s)", addr)
			}
			return nil
		},
		retry.DelayType(func(n uint, config *retry.Config) time.Duration {
			log.Printf("Retry connecting to redis (Addr:%s)", addr)
			return time.Second
		}),
		retry.Attempts(connRetryCount),
	); err != nil {
		return err
	}

	rc.client = client
	return nil
}

func (rc *RedisClient) Close() error {
	if err := rc.client.Close(); err != nil {
		return fmt.Errorf("cant close connect (%v)", err)
	}
	log.Print("close redis connection")
	return nil
}

func (rc *RedisClient) XRead(streamKeys []string, ids []string, count int, block time.Duration) ([]redis.XStream, error) {
	return rc.client.XRead(&redis.XReadArgs{
		Streams: append(streamKeys, ids...),
		Count:   int64(count),
		Block:   block,
	}).Result()
}

func (rc *RedisClient) XReadOne(streamKeys []string, ids []string, count int, block time.Duration) (map[string]interface{}, string, error) {
	ret, err := rc.client.XRead(&redis.XReadArgs{
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
	_, err := rc.client.XAdd(&redis.XAddArgs{
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
	_, err := rc.client.FlushAll().Result()
	return err
}

func (rc *RedisClient) XDel(streamKey string, ids []string) error {
	_, err := rc.client.XDel(streamKey, ids...).Result()
	return err
}

func (rc *RedisClient) HGet(key string) (map[string]string, error) {
	return rc.client.HGetAll(key).Result()
}

func (rc *RedisClient) HSet(key string, value map[string]interface{}) (int64, error) {
	count, err := rc.client.HSet(key, value).Result()
	return count, err
}

func (rc *RedisClient) Delete(keys ...string) (int64, error) {
	return rc.client.Del(keys...).Result()
}
