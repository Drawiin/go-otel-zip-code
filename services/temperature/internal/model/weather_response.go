package model

type CurrentWeather struct {
	TempC float64 `json:"temp_c"`
}

type WeatherResponse struct {
	Current CurrentWeather `json:"current"`
}
