package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"regexp"
	"zip-code-validation/config"
)

type ZipCodeRequest struct {
	Cep string `json:"cep"`
}

type TemperatureResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func validateZipCode(cep string) bool {
	re := regexp.MustCompile(`^\d{8}$`)
	return re.MatchString(cep)
}

func zipCodeHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ZipCodeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"message": "invalid zipcode"}`, http.StatusBadRequest)
			return
		}

		if !validateZipCode(req.Cep) {
			http.Error(w, `{"message": "invalid zipcode"}`, http.StatusUnprocessableEntity)
			return
		}

		resp, err := http.Get(fmt.Sprintf("%s/temperature/%s", cfg.TemperatureServiceURL, req.Cep))
		if err != nil {
			http.Error(w, `{"message": "error calling service B"}`, http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, `{"message": "error from service B"}`, resp.StatusCode)
			return
		}

		var tempResp TemperatureResponse
		if err := json.NewDecoder(resp.Body).Decode(&tempResp); err != nil {
			http.Error(w, `{"message": "error decoding response from service B"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tempResp)
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/zipcode", zipCodeHandler(cfg))

	http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), r)
}
