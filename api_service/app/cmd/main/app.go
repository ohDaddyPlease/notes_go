package main

import (
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"gitlab.konstweb.ru/ow/arch/notes/api_service/internal/client/category_service"
	"gitlab.konstweb.ru/ow/arch/notes/api_service/internal/client/note_service"
	"gitlab.konstweb.ru/ow/arch/notes/api_service/internal/client/tag_service"
	"gitlab.konstweb.ru/ow/arch/notes/api_service/internal/client/user_service"
	"gitlab.konstweb.ru/ow/arch/notes/api_service/internal/config"
	"gitlab.konstweb.ru/ow/arch/notes/api_service/internal/handlers/auth"
	"gitlab.konstweb.ru/ow/arch/notes/api_service/internal/handlers/categories"
	"gitlab.konstweb.ru/ow/arch/notes/api_service/internal/handlers/notes"
	"gitlab.konstweb.ru/ow/arch/notes/api_service/internal/handlers/tags"
	"gitlab.konstweb.ru/ow/arch/notes/pkg/cache/freecache"
	"gitlab.konstweb.ru/ow/arch/notes/pkg/jwt"
	"gitlab.konstweb.ru/ow/arch/notes/pkg/logging"
	"gitlab.konstweb.ru/ow/arch/notes/pkg/metrics"
	"gitlab.konstweb.ru/ow/arch/notes/pkg/shutdown"
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

	refreshTokenCache := freecache.NewCacheRepo(104857600) // 100MB

	jwtHelper := jwt.NewHelper(refreshTokenCache, logger)

	metricHandler := metric.Handler{Logger: logger}
	metricHandler.Register(router)

	userService := user_service.NewService(cfg.UserService.URL, "/users", logger)
	authHandler := auth.Handler{JWTHelper: jwtHelper, UserService: userService, Logger: logger}
	authHandler.Register(router)

	categoryService := category_service.NewService(cfg.CategoryService.URL, "/categories", logger)
	categoriesHandler := categories.Handler{CategoryService: categoryService, Logger: logger}
	categoriesHandler.Register(router)

	noteService := note_service.NewService(cfg.NoteService.URL, "/notes", logger)
	notesHandler := notes.Handler{NoteService: noteService, Logger: logger}
	notesHandler.Register(router)

	tagService := tag_service.NewService(cfg.TagService.URL, "/tags", logger)
	tagsHandler := tags.Handler{TagService: tagService, Logger: logger}
	tagsHandler.Register(router)

	logger.Println("start application")
	start(router, logger, cfg)
}

func start(router *httprouter.Router, logger logging.Logger, cfg *config.Config) {
	var server *http.Server
	var listener net.Listener

	if cfg.Listen.Type == "sock" {
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			logger.Fatal(err)
		}
		socketPath := path.Join(appDir, "app.sock")
		logger.Infof("socket path: %s", socketPath)

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

	logger.Println("application inited and started")

	if err := server.Serve(listener); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			logger.Warn("server shutdown")
		default:
			logger.Fatal(err)
		}
	}
}
