package main

import (
	"log"
	"net/http"
	"time"

	config "juansecalvinio/tepidolacuenta/config"
	database "juansecalvinio/tepidolacuenta/internal/database"
	middleware "juansecalvinio/tepidolacuenta/internal/middleware"
	pkg "juansecalvinio/tepidolacuenta/internal/pkg"

	authHandler "juansecalvinio/tepidolacuenta/internal/auth/handler"
	authRepo "juansecalvinio/tepidolacuenta/internal/auth/repository"
	authUseCase "juansecalvinio/tepidolacuenta/internal/auth/usecase"

	restaurantHandler "juansecalvinio/tepidolacuenta/internal/restaurant/handler"
	restaurantRepo "juansecalvinio/tepidolacuenta/internal/restaurant/repository"
	restaurantUseCase "juansecalvinio/tepidolacuenta/internal/restaurant/usecase"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize MongoDB
	db, err := database.NewMongoDB(cfg.MongoURI, "tepidolacuenta")
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}
	defer db.Close()
	log.Println("✓ Connected to MongoDB")

	// Initialize JWT service
	jwtService := pkg.NewJWTService(cfg.JWTSecret)

	// Initialize Auth module
	authRepository := authRepo.NewMongoRepository(db.Database)
	authService := authUseCase.NewAuthUseCase(authRepository, jwtService)
	authHdlr := authHandler.NewAuthHandler(authService)

	// Initialize Restaurant module
	restaurantRepository := restaurantRepo.NewMongoRepository(db.Database)
	restaurantService := restaurantUseCase.NewRestaurantUseCase(restaurantRepository)
	restaurantHdlr := restaurantHandler.NewRestaurantHandler(restaurantService)

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize Gin router
	router := gin.Default()

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.CORSAllowedOrigins},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now(),
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		authHdlr.RegisterRoutes(v1)

		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(jwtService))
		{
			// Restaurant routes
			restaurantHdlr.RegisterRoutes(protected)

			// Table routes - TODO: implement table module (using /tables path to avoid conflicts)
			protected.POST("/tables", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Create table endpoint - to be implemented"})
			})
			protected.GET("/tables/restaurant/:restaurantId", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Get tables by restaurant endpoint - to be implemented"})
			})
		}

		// Public routes (no authentication required)
		// Request routes - TODO: implement request module
		v1.POST("/request-account", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Request account endpoint - to be implemented"})
		})

		// WebSocket - TODO: implement websocket handler
		v1.GET("/ws/:restaurantId", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "WebSocket endpoint - to be implemented"})
		})
	}

	log.Printf("🚀 Server running on port %s\n", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
