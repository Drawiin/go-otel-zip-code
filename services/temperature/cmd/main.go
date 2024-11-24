package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go-zip-code-temperature/config"
	"go-zip-code-temperature/internal/client"
	"go-zip-code-temperature/internal/handler"
	"go-zip-code-temperature/internal/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	appConfig := getAppConfig()

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

	webClient := client.NewWebClient()
	temperatureService := getService(webClient, *appConfig, otelTracer)
	temperatureHandler := getHandler(temperatureService)
	server := newServer(ctx, appConfig, temperatureHandler)
	serverError := make(chan error, 1)
	go func() {
		log.Println("Starting server on port ", appConfig.Port)
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

func getService(client client.WebClient, config config.Config, tracer trace.Tracer) *service.CityTemperatureService {
	return service.NewCityTemperatureService(client, config, tracer)
}

func newServer(ctx context.Context, cfg *config.Config, temperatureHandler *handler.CityTemperatureHandler) *http.Server {
	// Set up the HTTP server - Chi router
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Get("/temperature/{cep}", temperatureHandler.GetTemperature)

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
