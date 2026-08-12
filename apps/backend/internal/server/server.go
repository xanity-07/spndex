package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/newrelic/go-agent/v3/integrations/nrredis-v9"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/xanity-07/spndex/internal/config"
	"github.com/xanity-07/spndex/internal/database"
	"github.com/xanity-07/spndex/internal/loggerpkg"
)

type Server struct {
	Config        *config.Config
	DB            *database.Database
	Redis         *redis.Client
	Logger        *zerolog.Logger
	LoggerService *loggerpkg.LoggerService
	httpServer    *http.Server
}

func New(cfg *config.Config, logger *zerolog.Logger, loggerService *loggerpkg.LoggerService) (*Server, error) {
	db, err := database.New(cfg, logger, loggerService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Redis client with New Relic integration
	opts, err := redis.ParseURL(cfg.Redis.Address)
	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(opts)
	// Add New Relic hook if available
	if loggerService != nil && loggerService.GetApplication() != nil {
		rdb.AddHook(nrredis.NewHook(rdb.Options()))
	}

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error().Err(err).Msg("Failed to connect to redis, continuing without Redis")
		// Dont fail startup if Redis is unabailable
	}
	logger.Info().Str("component", "Redis").Msg("connected to Redis")

	server := &Server{
		Config:        cfg,
		DB:            db,
		Redis:         rdb,
		Logger:        logger,
		LoggerService: loggerService,
	}

	// Start metrics collection
	// Runtime metrics are automatically collected by New Relic Go agent

	return server, nil
}

func (s *Server) SetupHTTPServer(handler http.Handler) {
	s.httpServer = &http.Server{
		Addr:         ":" + s.Config.Server.Port,
		Handler:      handler,
		ReadTimeout:  time.Duration(s.Config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.Config.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(s.Config.Server.IdleTimeout) * time.Second,
	}
}

func (s *Server) Start() error {
	if s.httpServer == nil {
		return errors.New("HTTP server not initialized")
	}

	s.Logger.Info().
		Str("port", s.Config.Server.Port).
		Str("env", s.Config.Primary.Env).
		Msg("starting server")

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	if err := s.Redis.Close(); err != nil {
		return fmt.Errorf("failed to shutdown Redus connection: %w", err)
	}

	s.Logger.Info().Msg("closing Redis connection")

	if err := s.DB.Close(); err != nil {
		return fmt.Errorf("failed to shutdown database connection: %w", err)
	}

	return nil
}
