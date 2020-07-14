package redispkg

import (
	"encoding/xml"
	"errors"
	"time"

	"github.com/go-redis/redis"
	"go.uber.org/zap"

	"github.com/overtalk/bgo/pkg/log"
	"github.com/overtalk/bgo/utils/xml"
)

type Config struct {
	XMLName xml.Name `xml:"xml"`
	Address struct {
		Item []string `xml:"item"`
	} `xml:"address"`
	Password string `xml:"password"`
	PoolSize int    `xml:"poolSize"`
}

type RedisClient struct {
	cfg    *Config
	client redis.Cmdable
}

func NewRedisClient(path string) (*RedisClient, error) {
	cfg := &Config{}
	if err := xmlutil.ParseXml(path, cfg); err != nil {
		return nil, err
	}

	return &RedisClient{cfg: cfg}, nil
}

func (this *RedisClient) Connect() error {
	if len(this.cfg.Address.Item) == 0 {
		return errors.New("redis address absent")
	}

	var conn redis.Cmdable
	if len(this.cfg.Address.Item) > 1 {
		conn = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    this.cfg.Address.Item,
			Password: this.cfg.Password,
			PoolSize: this.cfg.PoolSize,
		})
	} else {
		conn = redis.NewClient(&redis.Options{
			Addr:     this.cfg.Address.Item[0],
			Password: this.cfg.Password,
			PoolSize: this.cfg.PoolSize,
		})
	}

	if _, err := conn.Ping().Result(); err != nil {
		//logpkg.GetLogger().With(zap.Any("addrs", this.cfg.Address.Item)).Error("failed to ping redis during connection")
		logpkg.Error("failed to ping redis during connection", zap.Any("addrs", this.cfg.Address.Item))
		return err
	}

	this.client = conn
	return nil
}

func (this *RedisClient) GetConn() redis.Cmdable { return this.client }

func (this *RedisClient) Expire(key string, dur time.Duration) error {
	_, err := this.client.Expire(key, dur).Result()
	if err != nil {
		return err
	}
	return nil
}

func (this *RedisClient) Exist(key string) (bool, error) {
	count, err := this.client.Exists(key).Result()
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

func (this *RedisClient) Get(key string) (string, error) {
	return this.client.Get(key).Result()
}

func (this *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	return this.client.Set(key, value, expiration).Err()
}

func (this *RedisClient) INCR(key string) (int64, error) {
	return this.client.Incr(key).Result()
}

func (this *RedisClient) INCRBy(key string, value int64) (int64, error) {
	return this.client.IncrBy(key, value).Result()
}

func (this *RedisClient) HSet(key, field string, value interface{}, expiration time.Duration) error {
	if err := this.client.HSet(key, field, value).Err(); err != nil {
		return err
	}

	if err := this.client.Expire(key, expiration).Err(); err != nil {
		this.client.Del(key)
		return err
	}

	return nil
}

func (this *RedisClient) HMSet(key string, fields map[string]interface{}, expiration time.Duration) error {
	if err := this.client.HMSet(key, fields).Err(); err != nil {
		return err
	}

	if expiration > 0 {
		if err := this.client.Expire(key, expiration).Err(); err != nil {
			this.client.Del(key)
			return err
		}
	}

	return nil
}

func (this *RedisClient) HGet(key, field string) (string, error) {
	return this.client.HGet(key, field).Result()
}

func (this *RedisClient) HGetAll(key string) (map[string]string, error) {
	return this.client.HGetAll(key).Result()
}

func (this *RedisClient) Del(keys ...string) {
	this.client.Del(keys...)
}
