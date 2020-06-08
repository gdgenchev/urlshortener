package main

import (
	"encoding/json"
	"github.com/gdgenchev/urlshortener/internal/common/config"
	"github.com/gdgenchev/urlshortener/internal/urlshortener_service"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	configuration := readConfiguration()

	urlShortenerService := urlshortener_service.NewUrlShortenerService(configuration)
	defer urlShortenerService.ClosePersistenceManager()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/create", urlShortenerService.HandleGenerateShortSlug).Methods("POST")
	router.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/web/static/favicon.ico")
	})
	router.HandleFunc("/{short-slug}", urlShortenerService.HandleRedirectToRealUrl).Methods("GET")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/static/")))

	log.Fatal(http.ListenAndServe(":8080", router))
}

const configFile = "config/config.development.json"

func readConfiguration() config.Configuration {
	configPath, err := filepath.Abs(configFile)
	if err != nil {
		panic(err.Error())
	}

	file, err := os.Open(configPath)
	if err != nil {
		panic(err.Error())
	}
	defer closeFile(file)

	decoder := json.NewDecoder(file)
	var configuration config.Configuration

	err = decoder.Decode(&configuration)
	if err != nil {
		panic(err.Error())
	}

	return configuration
}

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		panic(err.Error())
	}
}
