package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Reverseflash45/sipinjam-lab/config"
	"github.com/Reverseflash45/sipinjam-lab/database"
	"github.com/Reverseflash45/sipinjam-lab/route"
)

func main() {
	config.LoadEnv()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	pool, err := database.NewPostgres(database.PostgresConfig{
		Host:     config.GetEnv("DB_HOST", "localhost"),
		Port:     config.GetEnv("DB_PORT", "5432"),
		User:     config.GetEnv("DB_USER", "postgres"),
		Password: config.GetEnv("DB_PASSWORD", ""),
		Name:     config.GetEnv("DB_NAME", "sipinjam_lab"),
		SSLMode:  config.GetEnv("DB_SSLMODE", "disable"),
		MaxConns: int32(config.GetEnvInt("DB_MAX_CONNS", 10)),
	})
	if err != nil {
		logger.Error("gagal terhubung ke PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	app := config.NewApp(logger, route.Dependencies{Pool: pool})

	go func() {
		port := config.GetEnv("APP_PORT", "3000")
		logger.Info("server berjalan", "port", port)
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti", "error", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	logger.Info("menghentikan server")
	_ = app.ShutdownWithTimeout(10 * time.Second)
}
