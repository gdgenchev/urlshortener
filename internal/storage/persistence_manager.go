//Package storage provides functionality for storing the data needed by the url shortener urlshortener_service.
package storage

import (
	"github.com/gdgenchev/urlshortener/internal/model"
	"github.com/gdgenchev/urlshortener/internal/util"
)

// PersistenceManager manages a long term database persistence and a short term
// cache persistence and provides thread safe methods for saving and retrieving data.
type PersistenceManager struct {
	databasePersistence DatabasePersistence
	cachePersistence    CachePersistence
}

func NewPersistenceManager(configuration util.Configuration) *PersistenceManager {
	persistenceManager := new(PersistenceManager)

	mysqlPersistence := NewMysqlPersistence(configuration)
	redisCachePersistence := NewRedisCachePersistence(configuration)

	persistenceManager.databasePersistence = mysqlPersistence
	persistenceManager.cachePersistence = redisCachePersistence

	return persistenceManager
}

// SaveUrlData persists the url data and returns false if the data already exists.
func (persistenceManager *PersistenceManager) SaveUrlData(urlData model.UrlData) bool {
	// If the data is present in the cache, we are sure that this is a duplicate
	if persistenceManager.cachePersistence.Exists(urlData.ShortSlug) {
		return false
	}

	// If the data has not been found in the cache, there is a chance that it is in the database
	// (if the cache memory limit has been reached and its eviction policy has been applied)
	if ok := persistenceManager.databasePersistence.SaveUrlData(urlData); !ok {
		return false
	}

	// The data has been inserted in the database, so we add it to the cache as well
	// Recently stored data = higher chance for url access
	persistenceManager.cachePersistence.SaveUrlData(urlData)
	return true
}

// GetRealUrl returns the real url given a short slug.
func (persistenceManager *PersistenceManager) GetRealUrl(shortSlug string) (string, bool) {
	// If the url data exists in the cache, we are sure that it is valid and return the real url
	realUrl, found := persistenceManager.cachePersistence.GetRealUrl(shortSlug)
	if found {
		return realUrl, true
	}

	// If the url data has not been found in the cache, it might be in the database, so we check.
	// If it is found in the database, we put it back in the cache as there is a high chance
	// that the url will be used in the near future.
	urlData, found := persistenceManager.databasePersistence.GetUrlData(shortSlug)
	if found {
		persistenceManager.cachePersistence.SaveUrlData(urlData)
		return urlData.RealUrl, true
	}

	return "", false
}

// Exists returns true if the short slug is already persisted in the cache or in the database.
func (persistenceManager *PersistenceManager) Exists(shortSlug string) bool {
	return persistenceManager.cachePersistence.Exists(shortSlug) ||
		persistenceManager.databasePersistence.Exists(shortSlug)
}

// Close closes the database persistence and the cache persistence.
func (persistenceManager *PersistenceManager) Close() {
	persistenceManager.databasePersistence.Close()
	persistenceManager.cachePersistence.Close()
}
