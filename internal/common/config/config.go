package config

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
