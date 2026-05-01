package app

import (
	"foxDenApp/internal/config"
	"foxDenApp/internal/logger/sl"
	"foxDenApp/internal/middleware"
	"foxDenApp/internal/services"
	"foxDenApp/internal/storage"
	"foxDenApp/internal/storage/repositories"
	"foxDenApp/internal/transport/http"

	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/gofiber/fiber/v2"

	_ "foxDenApp/docs"
)

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func Run(cfg *config.Config) {
	logger := sl.InitLogger(cfg.Env)

	logger.Info("Logger is enabled")
	logger.Debug("Debug is enabled")

	db := storage.Connect(cfg)

	logger.Info("Successfully connected to database!")

	storage := storage.Init(db)

	logger.Info("Successfully inited storage!")

	storage.Prepare()

	logger.Info("Successfully prepared db!")

	repos := repositories.InitRepositories(storage)

	logger.Info("Successfully inited repositories!")

	services := services.Init(repos, cfg)

	logger.Info("Successfully inited services!")

	app := fiber.New(fiber.Config{
		StrictRouting: true,
		WriteTimeout:  cfg.HTTPServer.Timeout,
		IdleTimeout:   cfg.HTTPServer.IdleTimeout,
	})
	app.Use(middleware.NewLogger(logger))
	swaggerCfg := swagger.Config{
		BasePath: "/api",
		FilePath: "./docs/swagger.json",
		Path:     "/docs",
		Title:    "Swagger API Docs",
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:8080",
		AllowCredentials: true,
	}))
	app.Use(swagger.New(swaggerCfg))

	http := http.Init(services, app)

	http.Start()
	app.Get("/docs/*", swagger.New(swaggerCfg))

	app.Listen(cfg.HTTPServer.Address)
}
