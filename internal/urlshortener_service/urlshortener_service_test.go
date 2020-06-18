package urlshortener_service_test

import (
	"bytes"
	"encoding/json"
	testing_utils "github.com/gdgenchev/urlshortener/internal/testing"
	"github.com/gdgenchev/urlshortener/internal/urlshortener_service"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const testRealUrl = "https://www.google.com/search?q=kittens&tbm=isch&ved=2ahUKEwj5_ZOS2IjqAhXRNuwKHeRzAVoQ2-cCegQIABAA&oq=kittens&gs_lcp=CgNpbWcQAzIECCMQJzICCAAyBAgAEB4yBAgAEB4yBAgAEB4yBAgAEB4yBAgAEB4yBAgAEB4yBAgAEB4yBAgAEB46BAgAEENQhwVYwgpgsgtoAHAAeACAAYIBiAHOBZIBAzMuNJgBAKABAaoBC2d3cy13aXotaW1n&sclient=img&ei=z_bpXrnaE9HtsAfk54XQBQ&bih=1164&biw=2327&rlz=1C1GCEB_enBG845BG845"
const testShortSlug = "kittens"

var testPersistence *testing_utils.TestPersistence
var urlShortenerService *urlshortener_service.UrlShortenerService

func TestMain(m *testing.M) {
	setUp()
	runTests := m.Run()
	cleanUp()
	os.Exit(runTests)
}

func cleanUp() {
	testPersistence.Close()
	urlShortenerService.ClosePersistenceManager()
}

func setUp() {
	testPersistence = testing_utils.NewTestPersistence()
	urlShortenerService = urlshortener_service.NewUrlShortenerService(testPersistence.GetTestConfiguration())
}

func sendRequestAndGetResponse(t *testing.T, jsonStr []byte) urlshortener_service.Response {
	req, err := http.NewRequest("POST", "/api/create", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(urlShortenerService.HandleGenerateShortSlug)
	handler.ServeHTTP(rr, req)

	responseBody, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatal(err)
	}

	var response urlshortener_service.Response

	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		t.Fatal(err)
	}

	return response
}

func TestCreateShortUrlWithOnlyRealUrlSpecified(t *testing.T) {
	testPersistence.FlushTestPersistence()

	var jsonStr = []byte(`{"real-url":"` + testRealUrl + `", "short-slug":"", "expires":""}`)
	response := sendRequestAndGetResponse(t, jsonStr)

	if response.ErrorMessage != "" {
		t.Errorf("HandleGenerateShortSlug returned an error: %v.\n", response.ErrorMessage)
	} else {
		t.Logf("Generated short url: %v", response.ShortUrl)
	}
}

func TestCreateShortUrlWithRealUrlAndUniqueShortSlugSpecified(t *testing.T) {
	testPersistence.FlushTestPersistence()

	var jsonStr = []byte(`{"real-url":"` + testRealUrl + `", "short-slug":"` + testShortSlug + `", "expires":""}`)
	response := sendRequestAndGetResponse(t, jsonStr)

	domain := testPersistence.GetTestConfiguration().UrlShortenerService.DomainName
	if response.ErrorMessage != "" {
		t.Errorf("HandleGenerateShortSlug returned an error: %v.\n", response.ErrorMessage)
	} else if response.ShortUrl != domain+"/"+testShortSlug {
		t.Errorf("HandleGenerateShortSlug returned a wrong short url: got %v want %v.\n",
			response.ShortUrl, domain+"/"+testShortSlug)
	}
}

func TestCreateShortUrlWithRealUrlAndDuplicateShortSlugSpecified(t *testing.T) {
	testPersistence.FlushTestPersistence()

	var jsonStr = []byte(`{"real-url":"` + testRealUrl + `", "short-slug":"` + testShortSlug + `", "expires":""}`)
	response := sendRequestAndGetResponse(t, jsonStr)
	response = sendRequestAndGetResponse(t, jsonStr)

	if response.ErrorMessage == "" {
		t.Errorf("Expected an error when trying to create a short url with a duplciate user defined slug")
	}
}

func TestHandleRedirectToRealUrlWithAValidShortSlug(t *testing.T) {
	testPersistence.FlushTestPersistence()

	var jsonStr = []byte(`{"real-url":"` + testRealUrl + `", "short-slug":"` + testShortSlug + `", "expires":""}`)
	response := sendRequestAndGetResponse(t, jsonStr)
	t.Log(response)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req = mux.SetURLVars(req, map[string]string{
		"short-slug": testShortSlug,
	})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(urlShortenerService.HandleRedirectToRealUrl)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMovedPermanently {
		t.Errorf("Expected a redirect status: %v, got status:%v.\n", http.StatusMovedPermanently, rr.Code)
	}
}

func TestCreateShortUrlWithAMissingRealUrl(t *testing.T) {
	testPersistence.FlushTestPersistence()

	var jsonStr = []byte(`{"real-url":"", "short-slug":"", "expires":""}`)
	response := sendRequestAndGetResponse(t, jsonStr)

	if response.ErrorMessage == "" {
		t.Errorf("Expected an error when sending an ill formatted request.\n")
	}
}

func TestCreateShortUrlWithAnIllFormattedRequest(t *testing.T) {
	testPersistence.FlushTestPersistence()

	var jsonStr = []byte(`{"real-url":", short-slug":, "expires":""}`)
	response := sendRequestAndGetResponse(t, jsonStr)

	if response.ErrorMessage == "" {
		t.Errorf("Expected an error when sending an ill formatted request.\n")
	}
}

func TestGetRealUrlWithANonExistentShortSlug(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req = mux.SetURLVars(req, map[string]string{
		"short-slug": testShortSlug,
	})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(urlShortenerService.HandleRedirectToRealUrl)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected a redirect status: %v, got status:%v.\n", http.StatusMovedPermanently, rr.Code)
	}
}
