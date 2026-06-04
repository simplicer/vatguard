package main

import (
	"log/slog"
	"net/http"
	"os"

	"codeberg.org/simplicer/vatguard/internal/vatvalidation/infrastructure"
)

var (
	Version   = "0.0.0"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logLevel := parseLogLevel(os.Getenv("LOG_LEVEL"))
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	logger := slog.New(handler)

	logger.Info("starting vatguard",
		slog.String("version", Version),
		slog.String("commit", Commit),
		slog.String("build_time", BuildTime),
		slog.String("port", port),
	)

	viesClient := infrastructure.NewVIESClient()
	router := infrastructure.NewRouter(viesClient, logger, Version)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	logger.Info("server listening", slog.String("addr", server.Addr))
	if err := server.ListenAndServe(); err != nil {
		logger.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "DEBUG", "debug":
		return slog.LevelDebug
	case "WARN", "warn":
		return slog.LevelWarn
	case "ERROR", "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
