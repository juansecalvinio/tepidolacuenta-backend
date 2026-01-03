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

	tableHandler "juansecalvinio/tepidolacuenta/internal/table/handler"
	tableRepo "juansecalvinio/tepidolacuenta/internal/table/repository"
	tableUseCase "juansecalvinio/tepidolacuenta/internal/table/usecase"

	requestDomain "juansecalvinio/tepidolacuenta/internal/request/domain"
	requestHandler "juansecalvinio/tepidolacuenta/internal/request/handler"
	requestRepo "juansecalvinio/tepidolacuenta/internal/request/repository"
	requestUseCase "juansecalvinio/tepidolacuenta/internal/request/usecase"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	// Initialize services
	jwtService := pkg.NewJWTService(cfg.JWTSecret)
	qrService := pkg.NewQRService(cfg.FrontendBaseURL)

	// Initialize WebSocket hub
	hub := pkg.NewHub()
	go hub.Run()
	log.Println("✓ WebSocket hub started")

	// Initialize Auth module
	authRepository := authRepo.NewMongoRepository(db.Database)
	authService := authUseCase.NewAuthUseCase(authRepository, jwtService)
	authHdlr := authHandler.NewAuthHandler(authService)

	// Initialize Restaurant module
	restaurantRepository := restaurantRepo.NewMongoRepository(db.Database)
	restaurantService := restaurantUseCase.NewRestaurantUseCase(restaurantRepository)
	restaurantHdlr := restaurantHandler.NewRestaurantHandler(restaurantService)

	// Initialize Table module
	tableRepository := tableRepo.NewMongoRepository(db.Database)
	tableService := tableUseCase.NewTableUseCase(tableRepository, restaurantRepository, qrService)
	tableHdlr := tableHandler.NewTableHandler(tableService)

	// Initialize Request module
	requestRepository := requestRepo.NewMongoRepository(db.Database)

	// Notification function for WebSocket
	notifyFunc := func(restaurantID primitive.ObjectID, request *requestDomain.Request) {
		hub.Broadcast(restaurantID, request)
	}

	requestService := requestUseCase.NewRequestUseCase(
		requestRepository,
		restaurantRepository,
		tableRepository,
		qrService,
		notifyFunc,
	)
	requestHdlr := requestHandler.NewRequestHandler(requestService, hub)

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

	// Public routes group (no authentication required)
	publicV1 := v1.Group("/public")

	{
		// Auth routes (public)
		authHdlr.RegisterRoutes(v1)

		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(jwtService))
		{
			// Restaurant routes
			restaurantHdlr.RegisterRoutes(protected)

			// Table routes
			tableHdlr.RegisterRoutes(protected)
		}

		// Request routes (both public and protected)
		requestHdlr.RegisterRoutes(protected, publicV1)
	}

	log.Printf("🚀 Server running on port %s\n", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
