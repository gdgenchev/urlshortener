package storage

import (
	"context"
	"encoding/json"
	"github.com/gdgenchev/urlshortener/internal/model"
	"github.com/gdgenchev/urlshortener/internal/util"
	"github.com/go-redis/redis/v8"
	"log"
	"strconv"
)

// TODO: Think whether it is really needed to fallback to a database, as the redis client takes too long
//  to realize that the redis server is down, which results in a very low performance.
//  I think the best decision for now is just to assume that the Redis cache server won't be down
//  or if it is down, we might have a switchover to another cache server?
//  This is overengineering at this point.

// CachePersistence provides a util interface for short term in memory url data persistence.
type CachePersistence interface {
	SaveUrlData(urlData model.UrlData)
	GetRealUrl(shortSlug string) (string, bool)
	Exists(shortSlug string) bool
	Close()
}

// RedisCachePersistence is a concrete implementation of CachePersistence
// If the cache is not running, the methods should log errors but not stop the complete execution
// because the application can fallback to a database only persistence.
type RedisCachePersistence struct {
	client *redis.Client
}

func NewRedisCachePersistence(configuration util.Configuration) *RedisCachePersistence {
	redisCachePersistence := new(RedisCachePersistence)

	redisCachePersistence.client = redis.NewClient(&redis.Options{
		Addr:     configuration.Redis.Host + ":" + strconv.Itoa(configuration.Redis.Port),
		Password: configuration.Redis.Password,
		DB:       configuration.Redis.DB,
	})

	return redisCachePersistence
}

// SaveUrlData saves the url data in the cache.
func (redisCachePersistence *RedisCachePersistence) SaveUrlData(urlData model.UrlData) {
	urlDataAsJson, err := json.Marshal(&urlData)
	if err != nil {
		log.Printf("Error in RedisCachePersistence.SaveUrlData(): %v.\n", err)
		return
	}

	redisCachePersistence.client.Set(context.Background(), urlData.ShortSlug, urlDataAsJson, 0)
	redisCachePersistence.client.ExpireAt(context.Background(), urlData.ShortSlug, urlData.Expires.Time)
}

// GetRealUrl retrieves the real url from the cache given a short slug.
func (redisCachePersistence *RedisCachePersistence) GetRealUrl(shortSlug string) (string, bool) {
	urlDataAsJson, err := redisCachePersistence.client.Get(context.Background(), shortSlug).Result()
	if err != nil {
		log.Printf("Error in RedisCachePersistence.GetRealUrl(): %v.\n", err)
		return "", false
	}

	var urlData model.UrlData
	err = json.Unmarshal([]byte(urlDataAsJson), &urlData)
	if err != nil {
		log.Printf("Error in RedisCachePersistence.GetRealUrl(): %v.\n", err)
		return "", false
	}

	return urlData.RealUrl, true
}

// Exists checks whether the short slug is present in the cache.
func (redisCachePersistence *RedisCachePersistence) Exists(shortSlug string) bool {
	exists, err := redisCachePersistence.client.Exists(context.Background(), shortSlug).Result()
	if err != nil {
		log.Printf("Error in RedisCachePersistence.Exists(): %v.\n", err)
		return false
	}
	return exists == 1
}

// Close closes the cache client.
func (redisCachePersistence *RedisCachePersistence) Close() {
	err := redisCachePersistence.client.Close()
	if err != nil {
		log.Printf("Error in RedisCachePersistence.Close(): %v.\n", err)
		return
	}
}
