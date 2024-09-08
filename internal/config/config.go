package config

import (
	"log"
	"os"
)

type Config struct {
    ServerPort string
    SecretKey string
}

// LoadConfig loads configuration from environment variables or defaults
func LoadConfig() (*Config, error) {
    port := os.Getenv("PORT")
    if port == "" {
        log.Println("No PORT environment variable set, using default 8080")
        port = "8080"
    }

    
    secret_key := os.Getenv("SECRET_KEY")
    if secret_key == "" {
        log.Println("No SECRET_KEY environment variable set, using default secret")
        secret_key = "secret"
    }

    return &Config{
        ServerPort: port,
        SecretKey: secret_key,
    }, nil
}
