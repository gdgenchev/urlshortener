package main

import (
	"github.com/gdgenchev/urlshortener/internal/urlshortener_service"
	"github.com/gdgenchev/urlshortener/internal/util"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const configFilePath = "config/config.development.json"

func main() {
	configuration := util.ReadConfiguration(configFilePath)

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
