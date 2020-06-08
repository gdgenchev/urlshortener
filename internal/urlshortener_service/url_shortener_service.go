// Package urlshortener_service provides the core logic and rest handlers for the url shortener.
package urlshortener_service

import (
	"encoding/json"
	"errors"
	"github.com/gdgenchev/urlshortener/internal/common/config"
	"github.com/gdgenchev/urlshortener/internal/common/urldata"
	"github.com/gdgenchev/urlshortener/internal/storage"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// response represents a json response.
type response struct {
	ShortUrl     string `json:"short-url"`
	ErrorMessage string `json:"error-message"`
}

// ShortSlugGenerator provides the logic for generating a short url slug.
type ShortSlugGenerator struct {
	SlugLength int
}

func NewShortSlugGenerator(slugLength int) *ShortSlugGenerator {
	shortSlugGenerator := new(ShortSlugGenerator)
	shortSlugGenerator.SlugLength = slugLength

	return shortSlugGenerator
}

// generateShortUrl generates a short slug by using a hardcoded alphabet.
// The algorithm is simple - choose N times a random symbol from the alphabet
// where N is equal to shortSlugGenerator.length. Extra persistence check should be
// made to make sure that the short url slug is unique.
func (shortSlugGenerator *ShortSlugGenerator) generateShortSlug() string {
	var alphabet = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	shortSlugBytes := make([]byte, shortSlugGenerator.SlugLength)
	randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := range shortSlugBytes {
		shortSlugBytes[i] = alphabet[randomGenerator.Intn(len(alphabet))]
	}

	return string(shortSlugBytes)
}

// UrlShortenerService wraps the REST handlers for the url shortener.
type UrlShortenerService struct {
	domainName         string
	defaultExpiresDays int
	shortSlugGenerator *ShortSlugGenerator
	persistenceManager *storage.PersistenceManager
	mutex              sync.Mutex
}

func NewUrlShortenerService(config config.Configuration) *UrlShortenerService {
	urlShortenerService := new(UrlShortenerService)

	urlShortenerService.domainName = config.UrlShortenerService.DomainName
	urlShortenerService.defaultExpiresDays = config.UrlShortenerService.DefaultExpireDays
	urlShortenerService.shortSlugGenerator = NewShortSlugGenerator(config.UrlShortenerService.SlugLength)
	urlShortenerService.persistenceManager = storage.NewPersistenceManager(config)

	return urlShortenerService
}

// HandleGenerateShortSlug is the REST handler for an incoming post request for creating a short url.
// There are 2 cases for handling the desired expire date of the short url:
// 	 1. The user has passed a desired expire date - then we persist that date.
// 	 2. The user has not passed a desired expire date - then we generate a default one - Now() + defaultExpiresDays
// There are 2 cases for handling the short url slug parameter:
// 	 1. The user has passed a desired short slug in the request
// 		- Then we just try to save it and if it fails, we return a high level error response, so as not to directly
//        inform the user for the existence of that short url.
// 	 2. The user has not passed a desired short slug(urlData.ShortSlug is equal to "")
// 		- Then we use the ShortSlugGenerator to generate a new random string and persist it.
func (urlShortenerService *UrlShortenerService) HandleGenerateShortSlug(writer http.ResponseWriter, request *http.Request) {
	urlData, err := urlShortenerService.getUrlDataFromRequest(request)
	if err != nil {
		urlShortenerService.sendErrorResponse(writer, http.StatusInternalServerError, "Error: Invalid Request")
		return
	}

	if urlData.Expires.IsZero() {
		urlData.Expires.Time = time.Now().Local().AddDate(0, 0, urlShortenerService.defaultExpiresDays)
	}

	urlShortenerService.mutex.Lock()
	if urlData.ShortSlug == "" {
		urlData.ShortSlug = urlShortenerService.generateUniqueShortSlug()
	}
	stored := urlShortenerService.persistenceManager.SaveUrlData(urlData)
	urlShortenerService.mutex.Unlock()

	if !stored {
		// Send a masked error message for the duplicate short slug, so as to provide some kind of protection :D
		urlShortenerService.sendErrorResponse(writer, http.StatusConflict,
			"Error: Please choose another short slug or leave it empty!")
		return
	}

	urlShortenerService.sendResponse(writer, http.StatusCreated,
		response{urlShortenerService.domainName + "/" + urlData.ShortSlug, ""})
}

// HandleRedirectToRealUrl is the REST handler for an incoming GET request for redirecting to the real url.
func (urlShortenerService *UrlShortenerService) HandleRedirectToRealUrl(writer http.ResponseWriter, request *http.Request) {
	shortSlug := mux.Vars(request)["short-slug"]

	realUrl, found := urlShortenerService.persistenceManager.GetRealUrl(shortSlug)

	if !found {
		urlShortenerService.sendErrorResponse(writer, http.StatusNotFound, "Error: URL Not Found")
		return
	}

	http.Redirect(writer, request, realUrl, http.StatusMovedPermanently)
}

// ClosePersistenceManager closes the open persistence services.
func (urlShortenerService *UrlShortenerService) ClosePersistenceManager() {
	urlShortenerService.persistenceManager.Close()
}

// Private helper methods
func (urlShortenerService *UrlShortenerService) getUrlDataFromRequest(request *http.Request) (urldata.UrlData, error) {
	var urlData urldata.UrlData

	reqBody, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error in getUrlDataFromRequest - ioutil.ReadAll(): %v", err)
		return urlData, errors.New("internal server error")
	}

	err = json.Unmarshal(reqBody, &urlData)
	if err != nil {
		log.Printf("Error in getUrlDataFromRequest() - json.Unmarshal(): %v\n", err)
		return urlData, errors.New("internal server error")
	}

	return urlData, nil
}

func (urlShortenerService *UrlShortenerService) generateUniqueShortSlug() string {
	shortSlug := urlShortenerService.shortSlugGenerator.generateShortSlug()
	for urlShortenerService.persistenceManager.Exists(shortSlug) {
		shortSlug = urlShortenerService.shortSlugGenerator.generateShortSlug()
	}

	return shortSlug
}

func (urlShortenerService *UrlShortenerService) sendErrorResponse(writer http.ResponseWriter, status int, errorMessage string) {
	urlShortenerService.sendResponse(writer, status, response{"", errorMessage})
}

func (urlShortenerService *UrlShortenerService) sendResponse(writer http.ResponseWriter, status int, response response) {
	writer.WriteHeader(status)
	err := json.NewEncoder(writer).Encode(response)
	if err != nil {
		log.Println("Error while encoding the response in json format: %v", err)
	}
}
