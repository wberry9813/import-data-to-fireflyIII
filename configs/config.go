package configs

import "os"

type Config struct {
	FireflyToken   string
	FireflyBaseURL string
}

func LoadConfig() Config {
	return Config{
		FireflyToken:   os.Getenv("FIREFLY_TOKEN"),
		FireflyBaseURL: os.Getenv("FIREFLY_BASE_URL"),
	}
}
