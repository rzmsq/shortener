package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"shortener/internal/config"
	httpController "shortener/internal/controller/http"
	"shortener/internal/storage/sqlite"
	"shortener/internal/usecase/shortener"
	"shortener/pkg/handler"
	"shortener/pkg/httpserver"
	"shortener/pkg/logger"
	"syscall"

	"github.com/go-playground/validator"
)

func Run(cfg *config.Config, log logger.Interface) error {
	var err error

	db, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init db", "err", err)
		return err
	}
	log.Info("db init success")

	useCase := shortener.NewUseCase(db)

	v := validator.New()
	controller := httpController.New(useCase, log, v)
	h := handler.New()
	h.AddHandler("POST /shorten", controller.Create)
	h.AddHandler("GET /s/{short_url}", controller.Open)
	h.AddHandler("GET /analytics/{short_url}", controller.Info)

	server := httpserver.New(h, cfg.Addr+":"+cfg.Port)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("Server starting...")
		if err = server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Failed serve", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shCtx, cancel := context.WithTimeout(ctx, cfg.MaxShutDownTimeServer)
	defer cancel()
	log.Info("Shutting down app")
	if err = server.Stop(shCtx); err != nil {
		log.Error("Error shutting down app", "error", err)
	}
	return err
}
