package config

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	Port        int           `yaml:"PORT" env-default:"8080"`
	JWTSecret   string        `yaml:"jwtsecret" env-default:"secret-code"`
	DBUser      string        `yaml:"dbuser" env-default:"postgres"`
	DBPassword  string        `yaml:"dbpassword" env-default:"postgres"`
	DBHost      string        `yaml:"dbhost" env-default:"db"`
	DBPort      string        `yaml:"dbport" env-default:"5432"`
	DBName      string        `yaml:"dbname" env-default:"avito_db"`
}

func LoadConfig() *Config {
	path := fetchConfigPath()
	if path == "" {
		log.Fatal("config path is empty")
	}

	return MustLoadByPath(path)
}

func MustLoadByPath(configPath string) *Config {
	if _, err := os.Stat(configPath); err != nil {
		log.Fatal("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatal("cannot read config: " + err.Error())
	}

	return &cfg
}

// fetchConfigPath fetches config path from command line flag or environment variable.
// Priority: flag > env > default.
// Default value is empty string.
func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
