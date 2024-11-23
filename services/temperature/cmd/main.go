package main

import (
	"fmt"
	"go-zip-code-temperature/config"
	"go-zip-code-temperature/internal/client"
	"go-zip-code-temperature/internal/handler"
	"go-zip-code-temperature/internal/service"
	"log"
	"net/http"
)

func main() {
	appConfig := getAppConfig()
	webClient := client.NewWebClient()
	temperatureService := getService(webClient, *appConfig)
	temperatureHandler := getHandler(temperatureService)
	port := appConfig.Port
	http.HandleFunc("GET /temperature/{cep}", temperatureHandler.GetTemperature)

	fmt.Println("Starting server on :" + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func getAppConfig() *config.Config {
	appConfig, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Error loading appConfig: %v", err)
	}
	return appConfig
}

func getHandler(service *service.CityTemperatureService) *handler.CityTemperatureHandler {
	return handler.NewCityTemperatureHandler(service)
}

func getService(client client.WebClient, config config.Config) *service.CityTemperatureService {
	return service.NewCityTemperatureService(client, config)
}
