package service_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go-zip-code-temperature/config"
	"go-zip-code-temperature/internal/model"
	"go-zip-code-temperature/internal/service"
)

// MockWebClient is a mock implementation of the WebClient interface
type MockWebClient struct {
	mock.Mock
}

func (m *MockWebClient) Get(url string) ([]byte, error) {
	args := m.Called(url)
	return args.Get(0).([]byte), args.Error(1)
}

func TestGetTemperatureSuccess(t *testing.T) {
	mockWebClient := new(MockWebClient)
	cfg := config.Config{
		CEPServiceURL: "http://cep-service",
		WeatherAPIURL: "http://weather-api",
		WeatherAPIKey: "test-key",
	}

	temperatureService := service.NewCityTemperatureService(mockWebClient, cfg)

	cep := "12345"
	cepUrl := fmt.Sprintf("%s/%s", cfg.CEPServiceURL, cep)
	weatherURL := fmt.Sprintf("%s?key=%s&q=%s&aqi=no", cfg.WeatherAPIURL, cfg.WeatherAPIKey, "testcity")

	mockWebClient.On("Get", cepUrl).Return([]byte(`{"city": "TéstCity"}`), nil)
	mockWebClient.On("Get", weatherURL).Return([]byte(`{"current": {"temp_c": 25.0}}`), nil)

	expectedResponse := model.TemperatureResponse{
		TempC: 25.0,
		TempF: 25.0*1.8 + 32,
		TempK: 25.0 + 273.15,
	}

	response, err := temperatureService.GetTemperature(cep)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
}

func TestGetTemperatureCEPServiceError(t *testing.T) {
	mockWebClient := new(MockWebClient)
	cfg := config.Config{
		CEPServiceURL: "http://cep-service",
		WeatherAPIURL: "http://weather-api",
		WeatherAPIKey: "test-key",
	}

	temperatureService := service.NewCityTemperatureService(mockWebClient, cfg)

	cep := "12345"
	cepUrl := fmt.Sprintf("%s/%s", cfg.CEPServiceURL, cep)

	mockWebClient.On("Get", cepUrl).Return([]byte{}, errors.New("CEP service error"))

	_, err := temperatureService.GetTemperature(cep)
	assert.Error(t, err)
}

func TestGetTemperatureWeatherServiceError(t *testing.T) {
	mockWebClient := new(MockWebClient)
	cfg := config.Config{
		CEPServiceURL: "http://cep-service",
		WeatherAPIURL: "http://weather-api",
		WeatherAPIKey: "test-key",
	}

	temperatureService := service.NewCityTemperatureService(mockWebClient, cfg)

	cep := "12345"
	cepUrl := fmt.Sprintf("%s/%s", cfg.CEPServiceURL, cep)
	weatherURL := fmt.Sprintf("%s?key=%s&q=%s&aqi=no", cfg.WeatherAPIURL, cfg.WeatherAPIKey, "testcity")

	mockWebClient.On("Get", cepUrl).Return([]byte(`{"city": "TéstCity"}`), nil)
	mockWebClient.On("Get", weatherURL).Return([]byte{}, errors.New("Weather service error"))

	_, err := temperatureService.GetTemperature(cep)
	assert.Error(t, err)
}
