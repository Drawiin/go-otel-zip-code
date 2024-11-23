package config

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io/fs"
)

var cfg *Config

type Config struct {
	CEPServiceURL string `mapstructure:"CEP_SERVICE_URL"`
	WeatherAPIURL string `mapstructure:"WEATHER_API_URL"`
	WeatherAPIKey string `mapstructure:"WEATHER_API_KEY"`
	Port          string `mapstructure:"PORT"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("app_config")
	viper.AddConfigPath(path)
	viper.SetConfigType("env")
	viper.SetConfigFile(".env")

	// Environment variables will override the .env values
	viper.AutomaticEnv()
	// Bind environment variables to the config struct if i dont do this it does not work
	viper.BindEnv("cep_service_url")
	viper.BindEnv("weather_api_url")
	viper.BindEnv("weather_api_key")
	viper.BindEnv("port")

	// Read the .env file
	if err := viper.ReadInConfig(); err != nil {
		// If the .env file is not found, continue without error
		fmt.Printf("Loading config from .env file %T\n", err)
		var pathError *fs.PathError
		if !errors.As(err, &pathError) {
			return nil, err
		}
	}

	err := viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
