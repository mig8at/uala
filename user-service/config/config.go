package config

import (
	"context"
	"fmt"
	"log"
	"user_service/internal/domain/models"

	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type Config struct {
	Port         string
	BadgerPath   string
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
		BadgerPath: viper.GetString("db.badger"),
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
	if err := db.AutoMigrate(&models.User{}, &models.Follower{}); err != nil {
		log.Fatalf("Error al migrar las tablas: %v", err)
	}
	db.Exec("PRAGMA foreign_keys = ON;")

	return db
}

func (c *Config) Redis() *redis.Client {
	rdb := redis.NewClient(c.RedisOptions)

	fmt.Println("Connecting to Redis...")
	fmt.Printf("Addr: %s\n", c.RedisOptions.Addr)
	fmt.Printf("Password: %s\n", c.RedisOptions.Password)
	fmt.Printf("DB: %d\n", c.RedisOptions.DB)

	// Optionally, test the connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Error connecting to Redis: %v", err)
	}
	return rdb
}
