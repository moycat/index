package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	adapterhttp "github.com/moycat/index/adapter/http"
	"github.com/moycat/index/adapter/index/ngram"
	"github.com/moycat/index/adapter/index/snippet"
	mysqlrepo "github.com/moycat/index/adapter/storage/mysql"
	"github.com/moycat/index/config"
	"github.com/moycat/index/service"
	log "github.com/sirupsen/logrus"
)

type App struct {
	Server *http.Server
	DB     *sql.DB
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{})
	if cfg.Debug {
		logger.SetLevel(log.DebugLevel)
	} else {
		logger.SetLevel(log.InfoLevel)
	}

	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open mysql connection: %w", err)
	}

	// Keep DB pool tuning simple and stable for initial deployment.
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	repo := mysqlrepo.NewPostRepository(db)
	indexService := service.NewIndexService(repo)
	searchService := service.NewSearchService(repo, ngram.NewTokenizer(2), snippet.NewBuilder(), cfg.SnippetMaxRunes)

	router := adapterhttp.NewRouter(adapterhttp.Dependencies{
		IndexService:  indexService,
		SearchService: searchService,
		AuthToken:     cfg.IndexToken,
		Debug:         cfg.Debug,
		Logger:        logger,
	})

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return &App{Server: server, DB: db}, nil
}

func (a *App) Close() error {
	if a.DB != nil {
		return a.DB.Close()
	}
	return nil
}
