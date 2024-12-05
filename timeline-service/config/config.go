package config

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

type Config struct {
	Port         string
	Env          string
	RedisOptions *redis.Options
}

func LoadConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error al leer la configuración: %v", err)
	}

	return &Config{
		Port: viper.GetString("server.port"),
		Env:  viper.GetString("env"),
		RedisOptions: &redis.Options{
			Addr:     viper.GetString("db.redis.addr"),
			Password: viper.GetString("db.redis.password"),
			DB:       viper.GetInt("db.redis.db"),
		},
	}
}
func (c *Config) Redis() *redis.Client {
	rdb := redis.NewClient(c.RedisOptions)

	// Optionally, test the connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Error connecting to Redis: %v", err)
	}
	return rdb
}
