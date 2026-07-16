package config

import "os"

type Config struct {
	Addr        string
	JWTSecret   string
	DatabaseDSN string
	RedisAddr   string
	RedisDB     int
}

func Load() Config {
	c := Config{Addr: ":8080", JWTSecret: "homeflow-development-secret", DatabaseDSN: "root:homeflow@tcp(mysql:3306)/homeflow?charset=utf8mb4&parseTime=True&loc=Local", RedisAddr: "redis:6379"}
	if value := os.Getenv("APP_ADDR"); value != "" {
		c.Addr = value
	}
	if value := os.Getenv("JWT_SECRET"); value != "" {
		c.JWTSecret = value
	}
	if value := os.Getenv("MYSQL_DSN"); value != "" {
		c.DatabaseDSN = value
	}
	if value := os.Getenv("REDIS_ADDR"); value != "" {
		c.RedisAddr = value
	}
	return c
}
