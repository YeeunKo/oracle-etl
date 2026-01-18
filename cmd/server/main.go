// Package main은 Oracle ETL API 서버의 진입점입니다
package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"oracle-etl/internal/adapter/handler"
	"oracle-etl/internal/adapter/sse"
	"oracle-etl/internal/config"
	"oracle-etl/internal/middleware"
	"oracle-etl/internal/repository/memory"
	"oracle-etl/internal/usecase"
)

const (
	// defaultConfigPath는 기본 설정 파일 경로입니다
	defaultConfigPath = "config.yaml"
)

func main() {
	// 로거 초기화 (JSON 포맷, UTC 타임스탬프)
	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Caller().
		Logger()

	zerolog.TimeFieldFormat = time.RFC3339

	// 설정 로드
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = defaultConfigPath
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("설정 로드 실패")
	}

	// 설정 유효성 검사
	if err := cfg.Validate(); err != nil {
		logger.Fatal().Err(err).Msg("설정 유효성 검사 실패")
	}

	logger.Info().
		Str("app", cfg.App.Name).
		Str("version", cfg.App.Version).
		Str("environment", cfg.App.Environment).
		Int("port", cfg.Server.Port).
		Msg("서버 시작 준비")

	// SSE Broadcaster 초기화
	broadcaster := sse.NewBroadcaster()

	// Broadcaster 컨텍스트 (서버 종료 시 함께 종료)
	broadcasterCtx, broadcasterCancel := context.WithCancel(context.Background())
	defer broadcasterCancel()

	// Broadcaster 시작
	go broadcaster.Run(broadcasterCtx)
	logger.Info().Msg("SSE Broadcaster 시작됨")

	// Repository 초기화 (In-Memory)
	transportRepo := memory.NewTransportRepository()
	jobRepo := memory.NewJobRepository()

	// Service 초기화
	transportSvc := usecase.NewTransportService(transportRepo)
	jobSvc := usecase.NewJobService(jobRepo, transportRepo)

	// Fiber 앱 초기화
	app := setupFiber(cfg, logger)

	// 라우트 설정
	setupRoutes(app, cfg, transportSvc, jobSvc, broadcaster)

	// 서버 시작 (goroutine)
	go func() {
		addr := ":" + strconv.Itoa(cfg.Server.Port)
		if err := app.Listen(addr); err != nil {
			logger.Fatal().Err(err).Msg("서버 시작 실패")
		}
	}()

	logger.Info().
		Int("port", cfg.Server.Port).
		Msg("서버 시작됨")

	// Graceful Shutdown 대기
	waitForShutdown(app, logger, broadcasterCancel)
}

// setupFiber는 Fiber 앱을 설정합니다
func setupFiber(cfg *config.Config, logger zerolog.Logger) *fiber.App {
	readTimeout, _ := time.ParseDuration(cfg.Server.ReadTimeout)
	writeTimeout, _ := time.ParseDuration(cfg.Server.WriteTimeout)

	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		// JSON 에러 핸들러
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"code":    "ERROR",
				"message": err.Error(),
			})
		},
	})

	// 미들웨어 적용
	app.Use(middleware.NewRecoveryMiddleware(logger))
	app.Use(middleware.NewLoggingMiddleware(logger))

	return app
}

// setupRoutes는 API 라우트를 설정합니다
func setupRoutes(app *fiber.App, cfg *config.Config, transportSvc *usecase.TransportService, jobSvc *usecase.JobService, broadcaster *sse.Broadcaster) {
	// Handlers 초기화
	healthHandler := handler.NewHealthHandler(cfg.App.Version)
	transportHandler := handler.NewTransportHandler(transportSvc, jobSvc)
	jobHandler := handler.NewJobHandler(jobSvc)
	statusHandler := handler.NewStatusHandler(broadcaster)

	// API 그룹
	api := app.Group("/api")

	// Health
	api.Get("/health", healthHandler.Check)

	// Transport CRUD
	api.Post("/transports", transportHandler.Create)
	api.Get("/transports", transportHandler.List)
	api.Get("/transports/:id", transportHandler.GetByID)
	api.Delete("/transports/:id", transportHandler.Delete)
	api.Post("/transports/:id/execute", transportHandler.Execute)

	// Transport 실시간 상태 (SSE)
	api.Get("/transports/:id/status", statusHandler.GetStatus)

	// Job 조회
	api.Get("/jobs", jobHandler.List)
	api.Get("/jobs/:id", jobHandler.GetByID)
}

// waitForShutdown은 종료 시그널을 대기하고 graceful shutdown을 수행합니다
func waitForShutdown(app *fiber.App, logger zerolog.Logger, broadcasterCancel context.CancelFunc) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info().Msg("종료 시그널 수신, 서버 종료 중...")

	// SSE Broadcaster 종료
	broadcasterCancel()
	logger.Info().Msg("SSE Broadcaster 종료됨")

	// 진행 중인 요청을 완료하기 위한 타임아웃
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error().Err(err).Msg("서버 종료 중 오류 발생")
	}

	logger.Info().Msg("서버 종료 완료")
}
