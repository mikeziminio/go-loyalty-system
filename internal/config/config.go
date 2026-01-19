package config

import (
	"flag"
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Address        string `envconfig:"RUN_ADDRESS"`
	DatabaseURI    string `envconfig:"DATABASE_URI"`
	AccrualAddress string `envconfig:"ACCRUAL_SYSTEM_ADDRESS"`
	LogLevel       string
}

var (
	DefaultAddress        = "localhost:8080"
	DefaultAccrualAddress = "http://localhost:8088"
	DefaultLogLevel       = "info"
)

func NewFromEnvsAndFlags() (*Config, error) {
	c := Config{}
	c.LogLevel = DefaultLogLevel

	flag.StringVar(&c.Address, "a", DefaultAddress, "host:port http сервера")
	flag.StringVar(&c.DatabaseURI, "d", "", "database URI")
	flag.StringVar(&c.AccrualAddress, "r", DefaultAccrualAddress, "http://host:port системы расчёта начислений")
	flag.Parse()

	// todo: исправить: переменные среды перезаписывают флаги
	err := envconfig.Process("", &c)
	if err != nil {
		return nil, fmt.Errorf("failed to process envs: %w", err)
	}

	return &c, nil
}
