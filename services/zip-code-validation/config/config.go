package config

import (
	"github.com/spf13/viper"
)

var cfg *Config

type Config struct {
	ServiceName           string `mapstructure:"SERVICE_NAME"`
	TemperatureServiceURL string `mapstructure:"TEMPERATURE_SERVICE_URL"`
	Port                  string `mapstructure:"PORT"`
}

func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()
	viper.BindEnv("SERVICE_NAME")
	viper.BindEnv("TEMPERATURE_SERVICE_URL")
	viper.BindEnv("PORT")

	err := viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
