package main

import (
	"escala-fds-api/internal/certificate"
	"escala-fds-api/internal/comment"
	"escala-fds-api/internal/holiday"
	"escala-fds-api/internal/plataform/database"
	"escala-fds-api/internal/swap"
	"escala-fds-api/internal/user"
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := database.NewMySQLConnection()
	if err != nil {
		logger.Fatal("database connection error", zap.Error(err))
	}

	// Repositories
	userRepo := user.NewRepository(db)
	swapRepo := swap.NewRepository(db)
	commentRepo := comment.NewRepository(db)
	holidayRepo := holiday.NewRepository(db)
	certificateRepo := certificate.NewRepository(db)

	// Services
	userService := user.NewService(userRepo)
	swapService := swap.NewService(swapRepo, userRepo, holidayRepo)
	commentService := comment.NewService(commentRepo, userRepo)
	holidayService := holiday.NewService(holidayRepo)
	certificateService := certificate.NewService(certificateRepo, userRepo)

	// Handlers
	userHandler := user.NewHandler(userService)
	swapHandler := swap.NewHandler(swapService)
	commentHandler := comment.NewHandler(commentService)
	holidayHandler := holiday.NewHandler(holidayService)
	certificateHandler := certificate.NewHandler(certificateService)

	// Router
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

	api := router.Group("/api")

	userHandler.RegisterRoutes(api)
	swapHandler.RegisterRoutes(api)
	commentHandler.RegisterRoutes(api)
	holidayHandler.RegisterRoutes(api)
	certificateHandler.RegisterRoutes(api)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info(fmt.Sprintf("server running on port %s", port))
	if err := router.Run(":" + port); err != nil {
		logger.Fatal("server run error", zap.Error(err))
	}
}
