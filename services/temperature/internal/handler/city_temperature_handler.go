package handler

import (
	"encoding/json"
	"go-zip-code-temperature/internal/service"
	"net/http"
)

type CityTemperatureHandler struct {
	service *service.CityTemperatureService
}

func NewCityTemperatureHandler(service *service.CityTemperatureService) *CityTemperatureHandler {
	return &CityTemperatureHandler{service: service}
}

func (h CityTemperatureHandler) GetTemperature(w http.ResponseWriter, r *http.Request) {
	cep := r.PathValue("cep")
	if cep == "" || len(cep) != 8 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid zipcode"))
		return
	}
	temperature, err := h.service.GetTemperature(cep)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("can not find zipcode"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(temperature)
}
