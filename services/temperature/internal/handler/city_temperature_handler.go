package handler

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go-zip-code-temperature/internal/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"net/http"
)

type CityTemperatureHandler struct {
	service *service.CityTemperatureService
}

func NewCityTemperatureHandler(service *service.CityTemperatureService) *CityTemperatureHandler {
	return &CityTemperatureHandler{service: service}
}

func (h CityTemperatureHandler) GetTemperature(w http.ResponseWriter, r *http.Request) {
	// Extract the context from the request, and inject it into the header for distributed tracing
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	cep := chi.URLParam(r, "cep")
	if cep == "" || len(cep) != 8 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid zipcode"))
		return
	}
	temperature, err := h.service.GetTemperature(ctx, cep)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("can not find zipcode"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(temperature)
}
