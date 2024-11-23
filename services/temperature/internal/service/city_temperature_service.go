package service

import (
	"encoding/json"
	"fmt"
	"go-zip-code-temperature/config"
	"go-zip-code-temperature/internal/client"
	"go-zip-code-temperature/internal/model"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"strings"
	"unicode"
)

type CityTemperatureService struct {
	webClient client.WebClient
	config    config.Config
}

func NewCityTemperatureService(webClient client.WebClient, config config.Config) *CityTemperatureService {
	return &CityTemperatureService{
		webClient: webClient,
		config:    config,
	}
}

func (s CityTemperatureService) GetTemperature(cep string) (model.TemperatureResponse, error) {
	fmt.Println("Reach out to CEP service at: ", s.config.CEPServiceURL)
	cepUrl := s.config.CEPServiceURL + "/" + cep
	cepResponse, err := s.webClient.Get(cepUrl)
	if err != nil {
		fmt.Println("Error reaching out to CEP service: ", err)
		return model.TemperatureResponse{}, err
	}
	fmt.Println("CEP response: ", string(cepResponse))

	address, err := toModel[model.AddressResponse](cepResponse)
	if err != nil {
		return model.TemperatureResponse{}, err
	}

	fmt.Println("Reach out to Weather service at: ", s.config.WeatherAPIURL)
	weatherURL := fmt.Sprintf("%s?key=%s&q=%s&aqi=no", s.config.WeatherAPIURL, s.config.WeatherAPIKey, sanitizeString(address.City))
	weatherResponse, err := s.webClient.Get(weatherURL)
	if err != nil {
		return model.TemperatureResponse{}, err
	}
	fmt.Println("Weather response: ", string(weatherResponse))

	weather, err := toModel[model.WeatherResponse](weatherResponse)
	if err != nil {
		return model.TemperatureResponse{}, err
	}

	return model.TemperatureResponse{
		City:  address.City,
		TempC: weather.Current.TempC,
		TempF: weather.Current.TempC*1.8 + 32,
		TempK: weather.Current.TempC + 273.15,
	}, nil
}

func toModel[T any](body []byte) (*T, error) {
	var modelStruct T
	err := json.Unmarshal(body, &modelStruct)
	if err != nil {
		return &modelStruct, err
	}
	return &modelStruct, nil
}

func sanitizeString(input string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	sanitized, _, _ := transform.String(t, input)

	return strings.ToLower(sanitized)
}
