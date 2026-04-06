package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	RunAddr          string `env:"RUN_ADDRESS" `
	DBURI            string `env:"DATABASE_URI"`
	AcrualAddr       string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	UserSecret       string `env:"USER_SECRET"`
	TokenExpireHours int    `env:"TOKEN_EXPIRE_HOURS"`
	WorkersNumber    int    `env:"WORKERS_NUMBER"`
}

func newConfigDefault() *Config {
	return &Config{
		RunAddr:          ":8080",
		UserSecret:       "UserTokenSecret",
		TokenExpireHours: 24 * 3,
		WorkersNumber:    5,
	}
}

// validate проверка заполненности обязательных параметров конфигурации
func (c *Config) validate() error {
	var errorText string
	switch {
	case c.DBURI == "":
		errorText = "не передана строка подключения к базе данных. Используйте флаг -d или переменную DATABASE_URI"
	case c.AcrualAddr == "":
		errorText = "не передан адрес системы расчёта начислений. Используйте флаг -r или переменную ACCRUAL_SYSTEM_ADDRESS"
	}
	if errorText != "" {
		return fmt.Errorf("%s", errorText)
	}
	return nil
}

// ParseConfig инициализация конфига приложения, чтение переменных из переменных окружения и переданных флагов
// значения переданных флагов являются более приоритетными, чем переменных окружения
func ParseConfig() (*Config, error) {
	cfg := newConfigDefault()

	err := env.Parse(cfg)
	if err != nil {
		return nil, fmt.Errorf("config env parse error %w", err)
	}

	flag.StringVar(&cfg.RunAddr, "a", cfg.RunAddr, "адрес и порт запуска сервиса")
	flag.StringVar(&cfg.DBURI, "d", cfg.DBURI, "адрес подключения к базе данных")
	flag.StringVar(&cfg.AcrualAddr, "r", cfg.AcrualAddr, "адрес системы расчёта начислений")
	flag.Parse()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
