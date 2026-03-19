package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/moycat/index/app"
	"github.com/moycat/index/config"
	log "github.com/sirupsen/logrus"
)

func main() {
	configPath := flag.String("config", "config.toml", "path to TOML config")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.WithError(err).Fatal("load config")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, err := app.New(ctx, cfg)
	if err != nil {
		log.WithError(err).Fatal("initialize app")
	}
	defer application.Close()

	go func() {
		if err := application.Server.ListenAndServe(); err != nil && err.Error() != "http: Server closed" {
			log.WithError(err).Fatal("http server failed")
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := application.Server.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Error("shutdown server")
		os.Exit(1)
	}
}
