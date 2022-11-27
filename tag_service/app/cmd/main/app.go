package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"gitlab.konstweb.ru/ow/arch/notes/tag_service/internal/config"
	"gitlab.konstweb.ru/ow/arch/notes/tag_service/internal/tag"
	"gitlab.konstweb.ru/ow/arch/notes/tag_service/internal/tag/db"
	"gitlab.konstweb.ru/ow/arch/notes/tag_service/pkg/handlers/metric"
	"gitlab.konstweb.ru/ow/arch/notes/tag_service/pkg/logging"
	mongo "gitlab.konstweb.ru/ow/arch/notes/tag_service/pkg/mongodb"
	"gitlab.konstweb.ru/ow/arch/notes/tag_service/pkg/shutdown"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	logging.Init()
	logger := logging.GetLogger()

	cfg := config.GetConfig()

	router := httprouter.New()

	metricHandler := metric.Handler{Logger: logger}
	metricHandler.Register(router)

	mongoClient, err := mongo.NewClient(context.Background(), cfg.MongoDB.Host, cfg.MongoDB.Port,
		cfg.MongoDB.Username, cfg.MongoDB.Password, cfg.MongoDB.Database, cfg.MongoDB.AuthDB)
	if err != nil {
		logger.Fatal(err)
	}
	tagStorage := db.NewStorage(mongoClient, cfg.MongoDB.Collection, logger)
	if err != nil {
		panic(err)
	}

	tagService, err := tag.NewService(tagStorage, logger)
	if err != nil {
		panic(err)
	}

	tagsHandler := tag.Handler{
		Logger:     logger,
		TagService: tagService,
	}
	tagsHandler.Register(router)

	logger.Println("start application")
	start(router, logger, cfg)
}

func start(router http.Handler, logger logging.Logger, cfg *config.Config) {
	var server *http.Server
	var listener net.Listener

	if cfg.Listen.Type == "sock" {
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			logger.Fatal(err)
		}
		socketPath := path.Join(appDir, "app.sock")
		logger.Infof("socket path: %s", socketPath)

		logger.Info("create and listen unix socket")
		listener, err = net.Listen("unix", socketPath)
		if err != nil {
			logger.Fatal(err)
		}
	} else {
		logger.Infof("bind application to host: %s and port: %s", cfg.Listen.BindIP, cfg.Listen.Port)

		var err error

		listener, err = net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Listen.BindIP, cfg.Listen.Port))
		if err != nil {
			logger.Fatal(err)
		}
	}

	server = &http.Server{
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go shutdown.Graceful([]os.Signal{syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM},
		server)

	logger.Println("application initialized and started")

	if err := server.Serve(listener); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			logger.Warn("server shutdown")
		default:
			logger.Fatal(err)
		}
	}
}
