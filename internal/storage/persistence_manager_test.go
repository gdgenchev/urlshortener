package storage_test

import (
	"github.com/gdgenchev/urlshortener/internal/model"
	"github.com/gdgenchev/urlshortener/internal/storage"
	testing_utils "github.com/gdgenchev/urlshortener/internal/testing"
	"os"
	"testing"
	"time"
)

var testPersistence *testing_utils.TestPersistence
var persistenceManager *storage.PersistenceManager
var testUrlData model.UrlData

func TestMain(m *testing.M) {
	setUp()
	runTests := m.Run()
	cleanUp()
	os.Exit(runTests)
}

func setUp() {
	testPersistence = testing_utils.NewTestPersistence()
	persistenceManager = storage.NewPersistenceManager(testPersistence.GetTestConfiguration())

	testShortSlug := "test-short-slug"
	testRealUrl := "http://very-long-real-url.com"
	testExpires := model.CustomTime{Time: time.Now().AddDate(0, 0, 1)}

	testUrlData = model.UrlData{ShortSlug: testShortSlug, RealUrl: testRealUrl, Expires: testExpires}
}

func cleanUp() {
	testPersistence.CleanUp()
	persistenceManager.Close()
}

func TestCreateNewShortUrlWhenUnique(t *testing.T) {
	testPersistence.FlushTestPersistence()

	ok := persistenceManager.SaveUrlData(testUrlData)

	if !ok {
		t.Errorf("Could not save url data - short slug already exists.")
	}
}

func TestCreateNewShortUrlWhenPresentInCache(t *testing.T) {
	testPersistence.FlushTestPersistence()

	persistenceManager.SaveUrlData(testUrlData)

	ok := persistenceManager.SaveUrlData(testUrlData)

	if ok {
		t.Errorf("Saved duplicate url data.")
	}
}

func TestCreateNewShortUrlWhenNotPresentInCacheButPresentInDatabase(t *testing.T) {
	testPersistence.FlushTestPersistence()

	persistenceManager.SaveUrlData(testUrlData)

	testPersistence.FlushTestCache()

	ok := persistenceManager.SaveUrlData(testUrlData)

	if ok {
		t.Errorf("Saved duplicate url data.")
	}
}

func TestGetRealUrlWhenPresentInCache(t *testing.T) {
	testPersistence.FlushTestPersistence()

	persistenceManager.SaveUrlData(testUrlData)

	foundRealUrl, found := persistenceManager.GetRealUrl(testUrlData.ShortSlug)
	if !found {
		t.Errorf("Real url: %s for short slug: %s was not found.", testUrlData.RealUrl, testUrlData.ShortSlug)
	} else {
		if foundRealUrl != testUrlData.RealUrl {
			t.Errorf("Expected real url: %s, got: %s.", testUrlData.RealUrl, foundRealUrl)
		}
	}
}

func TestGetRealUrlWhenNotPresentInCacheButPresentInDatabase(t *testing.T) {
	testPersistence.FlushTestPersistence()

	persistenceManager.SaveUrlData(testUrlData)

	testPersistence.FlushTestCache()

	foundRealUrl, found := persistenceManager.GetRealUrl(testUrlData.ShortSlug)
	if !found {
		t.Errorf("Real url: %s for short slug: %s  was not found.", testUrlData.RealUrl, testUrlData.ShortSlug)
	} else {
		if foundRealUrl != testUrlData.RealUrl {
			t.Errorf("Expected real url: %s, got: %s.", testUrlData.RealUrl, foundRealUrl)
		}
	}

	// Assert that the url data is now added to the cache as this is the logic in GetRealUrl
	if exists := testPersistence.ExistsInTestCache(testUrlData.ShortSlug); !exists {
		t.Errorf("The url data was not added to the cache after a cache miss and a database hit.")

	}
}

func TestGetRealUrlWhenNotPresentAnywhere(t *testing.T) {
	testPersistence.FlushTestPersistence()

	realUrl, found := persistenceManager.GetRealUrl(testUrlData.ShortSlug)
	if found {
		t.Errorf("GetRealUrl found inexistent real url: %s for short slug: %s", realUrl, testUrlData.ShortSlug)
	}
}

func TestExistsWhenPresentInCache(t *testing.T) {
	testPersistence.FlushTestPersistence()

	persistenceManager.SaveUrlData(testUrlData)

	exists := persistenceManager.Exists(testUrlData.ShortSlug)
	if !exists {
		t.Errorf("The url data for short slug: %s was not found.", testUrlData.ShortSlug)
	}
}

func TestExistsWhenNotPresentInCacheButPresentInDatabase(t *testing.T) {
	testPersistence.FlushTestPersistence()

	persistenceManager.SaveUrlData(testUrlData)

	testPersistence.FlushTestCache()

	exists := persistenceManager.Exists(testUrlData.ShortSlug)
	if !exists {
		t.Errorf("The url data for short slug: %s was not found.", testUrlData.ShortSlug)
	}
}
