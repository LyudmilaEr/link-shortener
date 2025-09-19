// TODO: init config: cleanenv
package main

// go run .\cmd\url-shortener

import (
	"log/slog"
	"os"

	"url-shortener/internal/config1"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/http-server/middleware/logger"
	"url-shortener/internal/lib/logger/handlers/slogpretty"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage/sqlite"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config1.MustLoad()
	log := setupLogger(cfg.Env)
	log.Info("starting url-shortener", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	// fmt.Printf("Env=%s\nStorage=%s\nHTTP=%+v\n", cfg.Env, cfg.StoragePath, cfg.HTTPServer)

	// TODO: init storage: sqlite
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	// // check
	// id, err := storage.SaveURL("https://www.google.com", "google")
	// if err != nil {
	// 	log.Error("failed to save url", sl.Err(err))
	// 	os.Exit(1)
	// }

	// log.Info("url saved", slog.Int64("id", id))

	// id, err = storage.SaveURL("https://www.google.com", "google")
	// if err != nil {
	// 	log.Error("failed to save url", sl.Err(err))
	// 	os.Exit(1)
	// }

	_ = storage

	// TODO: init router: chi, "chi render"
	router := chi.NewRouter()

	// middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/url", save.New(log, storage))

}

// конфигурация логгера
// slog - обёртка для логгера
func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyhandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}

// TODO: run server
