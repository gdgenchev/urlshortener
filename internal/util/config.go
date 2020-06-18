package util

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Configuration struct {
	Mysql struct {
		DriverName string
		Host       string
		Port       int
		User       string
		Password   string
		Database   string
	}

	Redis struct {
		Host     string
		Port     int
		Password string
		DB       int
	}

	UrlShortenerService struct {
		SlugLength        int
		DomainName        string
		DefaultExpireDays int
	}
}

func ReadConfiguration(configFilePath string) Configuration {
	configFile := openFile(configFilePath)
	defer closeFile(configFile)

	decoder := json.NewDecoder(configFile)
	var configuration Configuration

	err := decoder.Decode(&configuration)
	if err != nil {
		panic(err.Error())
	}

	return configuration
}

func openFile(configFilePath string) *os.File {
	configFilePath, err := filepath.Abs(configFilePath)
	if err != nil {
		panic(err.Error())
	}

	configFile, err := os.Open(configFilePath)
	if err != nil {
		panic(err.Error())
	}

	return configFile
}

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		panic(err.Error())
	}
}
