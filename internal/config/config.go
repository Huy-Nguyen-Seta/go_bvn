package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	RedisHost         string
	RedisPort         string
	ElasticsearchURL  string
	Port              string
}

func Load() *Config {
	return &Config{
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBUser:           getEnv("DB_USER", "huybvn"),
		DBPassword:       getEnv("DB_PASSWORD", "huybvnpass"),
		DBName:           getEnv("DB_NAME", "blog_db"),
		RedisHost:        getEnv("REDIS_HOST", "localhost"),
		RedisPort:        getEnv("REDIS_PORT", "6379"),
		ElasticsearchURL: getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
		Port:             getEnv("PORT", "8080"),
	}
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}