package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go-zip-code-temperature/config"
	"go-zip-code-temperature/internal/client"
	"go-zip-code-temperature/internal/model"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"strings"
	"unicode"
)

type CityTemperatureService struct {
	webClient client.WebClient
	config    config.Config
	tracer    trace.Tracer
}

func NewCityTemperatureService(webClient client.WebClient, config config.Config, tracer trace.Tracer) *CityTemperatureService {
	return &CityTemperatureService{
		webClient: webClient,
		config:    config,
		tracer:    tracer,
	}
}

func (s CityTemperatureService) GetTemperature(ctx context.Context, cep string) (model.TemperatureResponse, error) {
	ctx, span := s.tracer.Start(ctx, "get temperature")
	span.SetAttributes(attribute.String("zip-code", cep))
	defer span.End()

	cepUrl := s.config.CEPServiceURL + "/" + cep
	ctx, spanCep := s.tracer.Start(ctx, "get address")
	spanCep.SetAttributes(attribute.String("url", cepUrl))
	cepResponse, err := s.webClient.Get(cepUrl)
	if err != nil {
		spanCep.RecordError(err)
		spanCep.SetStatus(codes.Error, "Failed to get address")
		spanCep.End()
		return model.TemperatureResponse{}, err
	}
	spanCep.SetStatus(codes.Ok, "Success")
	spanCep.End()

	address, err := toModel[model.AddressResponse](cepResponse)
	if err != nil {
		return model.TemperatureResponse{}, err
	}

	ctx, spanTemp := s.tracer.Start(ctx, "get weather")
	weatherURL := fmt.Sprintf("%s?key=%s&q=%s&aqi=no", s.config.WeatherAPIURL, s.config.WeatherAPIKey, sanitizeString(address.City))
	spanTemp.SetAttributes(attribute.String("url", weatherURL))
	weatherResponse, err := s.webClient.Get(weatherURL)
	if err != nil {
		spanTemp.RecordError(err)
		spanTemp.SetStatus(codes.Error, "Failed to get weather")
		spanTemp.End()
		return model.TemperatureResponse{}, err
	}
	spanTemp.SetStatus(codes.Ok, "Success")
	spanTemp.End()

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
