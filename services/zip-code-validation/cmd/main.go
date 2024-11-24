package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"time"
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

const name = "console"

var (
	tracer = otel.Tracer(name)
	logger = otelslog.NewLogger(name)
)

func validateZipCode(cep string) bool {
	re := regexp.MustCompile(`^\d{8}$`)
	return re.MatchString(cep)
}

func zipCodeHandler(cfg *config.Config, otelTracer trace.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the context from the request, and inject it into the header for distributed tracing
		carrier := propagation.HeaderCarrier(r.Header)
		ctx := r.Context()
		ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

		ctx, span := otelTracer.Start(ctx, "validateZipCode - zip-code-validation")
		defer span.End()

		var req ZipCodeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.InfoContext(ctx, err.Error(), "result", "Invalid zip code")
			http.Error(w, `{"message": "invalid zipcode"}`, http.StatusBadRequest)
			return
		}

		span.SetAttributes(attribute.String("zip-code", req.Cep))
		if !validateZipCode(req.Cep) {
			logger.InfoContext(ctx, "Invalid zip code")
			http.Error(w, `{"message": "invalid zipcode"}`, http.StatusUnprocessableEntity)
			return
		}

		// Create a new context to call the temperature service
		request, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/temperature/%s", cfg.TemperatureServiceURL, req.Cep), nil)
		if err != nil {
			http.Error(w, `{"message": "error calling service B"}`, http.StatusInternalServerError)
			return
		}

		// Execute the request
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(request.Header))
		resp, err := http.DefaultClient.Do(request)
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
	// Load the configuration.
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Set up OpenTelemetry.
	otelShutdown, err := setupOTelSDK(ctx)
	if err != nil {
		return
	}

	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	otelTracer := otel.Tracer("microservice-tracer")

	// Set up the HTTP server
	server := newServer(cfg, ctx, otelTracer)
	serverError := make(chan error, 1)
	go func() {
		log.Println("Starting server on port ", cfg.Port)
		serverError <- server.ListenAndServe()
	}()

	// Wait for interruption.
	select {
	case err = <-serverError:
		// Error when starting HTTP server.
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	err = server.Shutdown(context.Background())
	return
}

func newServer(cfg *config.Config, ctx context.Context, otelTracer trace.Tracer) *http.Server {
	// Set up the HTTP server - Chi router
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Post("/zipcode", zipCodeHandler(cfg, otelTracer))

	// Set up the HTTP server - Server
	// Set up this way to be able to shut down the server gracefully.
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      router,
	}
	return server
}
