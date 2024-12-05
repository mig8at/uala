package config

import (
	"log"
	"tweet-service/internal/domain/models"

	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type Config struct {
	Port         string
	SqlitePath   string
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
		Port:       viper.GetString("server.port"),
		SqlitePath: viper.GetString("db.sqlite"),
		Env:        viper.GetString("env"),
		RedisOptions: &redis.Options{
			Addr:     viper.GetString("db.redis.addr"),
			Password: viper.GetString("db.redis.password"),
			DB:       viper.GetInt("db.redis.db"),
		},
	}
}

func (c *Config) Sqlite() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(c.SqlitePath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error al conectar con la base de datos SQLite: %v", err)
	}

	// Migrar los modelos para crear tablas automáticamente
	if err := db.AutoMigrate(&models.Tweet{}, &models.Tag{}, &models.Comment{}); err != nil {
		log.Fatalf("Error al migrar las tablas: %v", err)
	}
	db.Exec("PRAGMA foreign_keys = ON;")

	return db
}

func (c *Config) Redis() *redis.Client {
	rdb := redis.NewClient(c.RedisOptions)

	return rdb
}
