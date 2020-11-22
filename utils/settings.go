package utils

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"log"
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Print("No .env file found")
	} else {
		return
	}
	if err := godotenv.Load("../.env"); err != nil {
		log.Print("No ../.env file found")
	}
}

type Settings struct {
	ListenPort int    `env:"LISTEN_PORT"`
	DbPath     string `env:"DB_PATH"`
	LogRequest bool `env:"LOG_REQUEST"`
	LogResponse bool `env:"LOG_RESPONSE"`
}

func ReadSettings() *Settings {
	var cfg Settings
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Can't parse config: %v", err)
	}
	return &cfg
}
