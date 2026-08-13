package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xanity-07/spndex/internal/config"
	"github.com/xanity-07/spndex/internal/database"
	"github.com/xanity-07/spndex/internal/handlers"
	"github.com/xanity-07/spndex/internal/loggerpkg"
	"github.com/xanity-07/spndex/internal/repositories"
	"github.com/xanity-07/spndex/internal/router"
	"github.com/xanity-07/spndex/internal/server"
	"github.com/xanity-07/spndex/internal/service"
)

const (
	DefaultContextTimeout = 30
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	gin.SetMode(gin.ReleaseMode)

	// Initialize New Relic logger service
	loggerService := loggerpkg.NewLoggerService(cfg.Observability)
	defer loggerService.Shutdown()

	log := loggerpkg.NewLoggerWithService(cfg.Observability, loggerService)

	if cfg.Primary.Env == "local" {
		if err = database.Migrate(context.Background(), cfg, &log); err != nil {
			log.Fatal().Err(err).Msg("failed to migrate database")
		}
	}

	// Initialize server
	srv, err := server.New(cfg, &log, loggerService)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize server")
	}

	// Initialize repositories, services, and handlers
	repos := repositories.NewRepositories(srv)
	services := service.NewServices(srv, repos)
	handlers := handlers.NewHandlers(srv, services)

	// Initialize router
	r := router.NewRouter(srv, handlers)

	// Setup HTTP server
	srv.SetupHTTPServer(r)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	// Start server
	go func() {
		if err = srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	// Wait for interruption signal to gracefully shutdown the server
	// 30 second timeout allows for any currently ongoing request to get processed
	// but any new incomming requests will get dropped
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultContextTimeout*time.Second)

	if err = srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}
	stop()
	cancel()

	log.Info().Msg("server exited properly")
}
