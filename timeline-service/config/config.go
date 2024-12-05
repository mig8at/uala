package config

import (
	"log"

	"github.com/dgraph-io/badger/v4"
	"github.com/spf13/viper"
)

type Config struct {
	Port       string
	BadgerPath string
	Env        string
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
		Env:        viper.GetString("env"),
	}
}

func (c *Config) Badger() *badger.DB {
	opts := badger.DefaultOptions("").WithInMemory(true).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}

	return db
}
