package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Env         string
	DatabaseURL string
	JWTSecret   string
	Port        string
}

var C AppConfig

func Load() {
	env := os.Getenv("GO_ENV")
	file := ".env"
	if env == "test" {
		file = ".env.test"
	}
	_ = godotenv.Load(file)

	C = AppConfig{
		Env:         ifEmpty(env, "dev"),
		DatabaseURL: mustEnv("DATABASE_URL"),
		JWTSecret:   mustEnv("JWT_SECRET"),
		Port:        ifEmpty(os.Getenv("PORT"), "8080"),
	}
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing env: %s", k)
	}
	return v
}
func ifEmpty(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
