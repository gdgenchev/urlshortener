package storage

import (
	"context"
	"encoding/json"
	"github.com/gdgenchev/urlshortener/internal/common/config"
	"github.com/gdgenchev/urlshortener/internal/common/urldata"
	"github.com/go-redis/redis/v8"
	"strconv"
)

// CachePersistence provides a common interface for short term in memory url data persistence.
type CachePersistence interface {
	SaveUrlData(urlData urldata.UrlData)
	GetRealUrl(shortSlug string) (string, bool)
	Exists(shortSlug string) bool
	Close()
}

// RedisCachePersistence is a concrete implementation of CachePersistence
type RedisCachePersistence struct {
	client *redis.Client
}

func NewRedisCachePersistence(configuration config.Configuration) *RedisCachePersistence {
	redisCachePersistence := new(RedisCachePersistence)

	redisCachePersistence.client = redis.NewClient(&redis.Options{
		Addr:     configuration.Redis.Host + ":" + strconv.Itoa(configuration.Redis.Port),
		Password: configuration.Redis.Password,
		DB:       configuration.Redis.DB,
	})

	return redisCachePersistence
}

// SaveUrlData saves the url data in the cache.
func (redisCachePersistence *RedisCachePersistence) SaveUrlData(urlData urldata.UrlData) {
	urlDataAsJson, err := json.Marshal(&urlData)
	if err != nil {
		panic(err.Error())
	}

	redisCachePersistence.client.Set(context.Background(), urlData.ShortSlug, urlDataAsJson, 0)
	redisCachePersistence.client.ExpireAt(context.Background(), urlData.ShortSlug, urlData.Expires.Time)
}

// GetRealUrl retrieves the real url from the cache given a short slug.
func (redisCachePersistence *RedisCachePersistence) GetRealUrl(shortSlug string) (string, bool) {
	urlDataAsJson, err := redisCachePersistence.client.Get(context.Background(), shortSlug).Result()
	if err == redis.Nil {
		return "", false
	}

	var urlData urldata.UrlData
	err = json.Unmarshal([]byte(urlDataAsJson), &urlData)
	if err != nil {
		return "", false
	}

	return urlData.RealUrl, true
}

// Exists checks whether the short slug is present in the cache.
func (redisCachePersistence *RedisCachePersistence) Exists(shortSlug string) bool {
	exists, err := redisCachePersistence.client.Exists(context.Background(), shortSlug).Result()
	if err != nil {
		panic(err.Error())
	}
	return exists == 1
}

// Close closes the cache client.
func (redisCachePersistence *RedisCachePersistence) Close() {
	err := redisCachePersistence.client.Close()
	if err != nil {
		panic(err.Error())
	}
}
