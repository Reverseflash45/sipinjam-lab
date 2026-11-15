package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Reverseflash45/sipinjam-lab/app/repository"
	"github.com/Reverseflash45/sipinjam-lab/app/service"
	"github.com/Reverseflash45/sipinjam-lab/config"
	"github.com/Reverseflash45/sipinjam-lab/database"
	"github.com/Reverseflash45/sipinjam-lab/route"
)

func main() {
	config.LoadEnv()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < 32 {
		logger.Error("JWT_SECRET wajib diisi minimal 32 karakter")
		os.Exit(1)
	}

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

	// --- Perakitan dependency: repository -> service ---
	userRepo := repository.NewUserRepository(pool)
	roomRepo := repository.NewRoomRepository(pool)
	itemRepo := repository.NewItemRepository(pool)
	bookingRepo := repository.NewBookingRepository(pool)
	txm := repository.NewTxManager(pool)

	authService := service.NewAuthService(
		userRepo,
		jwtSecret,
		time.Duration(config.GetEnvInt("JWT_TTL_MINUTES", 60))*time.Minute,
	)

	app := config.NewApp(logger, route.Dependencies{
		Pool:           pool,
		AuthService:    authService,
		JWTSecret:      jwtSecret,
		RoomService:    service.NewRoomService(roomRepo, bookingRepo),
		ItemService:    service.NewItemService(itemRepo),
		BookingService: service.NewBookingService(bookingRepo, roomRepo, itemRepo, txm),
	})

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
