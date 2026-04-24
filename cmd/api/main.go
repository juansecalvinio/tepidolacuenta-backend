package main

import (
	"context"
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

	invitationHandler "juansecalvinio/tepidolacuenta/internal/invitation/handler"
	invitationRepo "juansecalvinio/tepidolacuenta/internal/invitation/repository"
	invitationUseCase "juansecalvinio/tepidolacuenta/internal/invitation/usecase"

	restaurantHandler "juansecalvinio/tepidolacuenta/internal/restaurant/handler"
	restaurantRepo "juansecalvinio/tepidolacuenta/internal/restaurant/repository"
	restaurantUseCase "juansecalvinio/tepidolacuenta/internal/restaurant/usecase"

	branchHandler "juansecalvinio/tepidolacuenta/internal/branch/handler"
	branchRepo "juansecalvinio/tepidolacuenta/internal/branch/repository"
	branchUseCase "juansecalvinio/tepidolacuenta/internal/branch/usecase"

	tableHandler "juansecalvinio/tepidolacuenta/internal/table/handler"
	tableRepo "juansecalvinio/tepidolacuenta/internal/table/repository"
	tableUseCase "juansecalvinio/tepidolacuenta/internal/table/usecase"

	requestDomain "juansecalvinio/tepidolacuenta/internal/request/domain"
	requestHandler "juansecalvinio/tepidolacuenta/internal/request/handler"
	requestRepo "juansecalvinio/tepidolacuenta/internal/request/repository"
	requestUseCase "juansecalvinio/tepidolacuenta/internal/request/usecase"

	"juansecalvinio/tepidolacuenta/internal/migration"

	setupHandler "juansecalvinio/tepidolacuenta/internal/setup/handler"
	setupUseCase "juansecalvinio/tepidolacuenta/internal/setup/usecase"

	subscriptionDomain "juansecalvinio/tepidolacuenta/internal/subscription/domain"
	subscriptionHandler "juansecalvinio/tepidolacuenta/internal/subscription/handler"
	subscriptionRepo "juansecalvinio/tepidolacuenta/internal/subscription/repository"
	subscriptionUseCase "juansecalvinio/tepidolacuenta/internal/subscription/usecase"

	paymentDomain "juansecalvinio/tepidolacuenta/internal/payment/domain"
	paymentHandler "juansecalvinio/tepidolacuenta/internal/payment/handler"
	mpInfra "juansecalvinio/tepidolacuenta/internal/payment/infrastructure/mercadopago"
	paymentRepo "juansecalvinio/tepidolacuenta/internal/payment/repository"
	paymentUseCase "juansecalvinio/tepidolacuenta/internal/payment/usecase"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Sentry
	if cfg.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.SentryDSN,
			EnableTracing:    true,
			TracesSampleRate: 1.0,
			EnableLogs:       true,
		}); err != nil {
			log.Printf("Sentry initialization failed: %v", err)
		} else {
			defer sentry.Flush(2 * time.Second)
			log.Println("✓ Sentry initialized")
		}
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
	emailService := pkg.NewEmailService(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFrom)

	// Initialize WebSocket hub
	hub := pkg.NewHub()
	go hub.Run()
	log.Println("✓ WebSocket hub started")

	// Initialize Google OAuth config
	googleOAuth := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	// Initialize Subscription repositories early (needed by restaurant usecase)
	planRepository := subscriptionRepo.NewMongoPlanRepository(db.Database)
	subscriptionRepository := subscriptionRepo.NewMongoSubscriptionRepository(db.Database)

	// Initialize Restaurant module
	restaurantRepository := restaurantRepo.NewMongoRepository(db.Database)
	restaurantService := restaurantUseCase.NewRestaurantUseCase(restaurantRepository, planRepository, subscriptionRepository)
	restaurantHdlr := restaurantHandler.NewRestaurantHandler(restaurantService)

	// Initialize Invitation module
	invitationRepository := invitationRepo.NewMongoRepository(db.Database)
	invitationService := invitationUseCase.NewInvitationUseCase(invitationRepository, restaurantRepository)
	invitationHdlr := invitationHandler.NewInvitationHandler(invitationService)

	// Initialize Auth module
	authRepository := authRepo.NewMongoRepository(db.Database)
	authService := authUseCase.NewAuthUseCase(authRepository, jwtService, emailService, cfg.FrontendBaseURL, googleOAuth, invitationService)
	authHdlr := authHandler.NewAuthHandler(authService, cfg.JWTSecret, cfg.FrontendBaseURL)

	// Initialize Branch module
	branchRepository := branchRepo.NewMongoRepository(db.Database)
	branchService := branchUseCase.NewBranchUseCase(branchRepository, restaurantRepository, subscriptionRepository, planRepository)
	branchHdlr := branchHandler.NewBranchHandler(branchService)

	// Initialize Table module
	tableRepository := tableRepo.NewMongoRepository(db.Database)
	tableService := tableUseCase.NewTableUseCase(tableRepository, branchRepository, restaurantRepository, subscriptionRepository, planRepository, qrService)
	tableHdlr := tableHandler.NewTableHandler(tableService)

	// Initialize Setup module
	setupService := setupUseCase.NewSetupUseCase(restaurantRepository, branchRepository, tableRepository, qrService)
	setupHdlr := setupHandler.NewSetupHandler(setupService)

	// Initialize Subscription module
	subscriptionService := subscriptionUseCase.NewSubscriptionUseCase(planRepository, subscriptionRepository, restaurantRepository)
	subscriptionHdlr := subscriptionHandler.NewSubscriptionHandler(subscriptionService)

	// Initialize Payment module
	mpClient, err := mpInfra.NewClient(cfg.MercadoPagoAccessToken, cfg.MercadoPagoWebhookSecret)
	if err != nil {
		log.Fatalf("Failed to initialize MercadoPago client: %v", err)
	}
	paymentRepository := paymentRepo.NewMongoRepository(db.Database)
	notifyPaymentFunc := func(restaurantID primitive.ObjectID, event *paymentDomain.PaymentEvent) {
		hub.Broadcast(restaurantID, event)
	}
	paymentService := paymentUseCase.NewPaymentUseCase(paymentRepository, planRepository, subscriptionRepository, restaurantRepository, mpClient, cfg.MercadoPagoNotificationURL, cfg.FrontendBaseURL, notifyPaymentFunc)
	paymentHdlr := paymentHandler.NewPaymentHandler(paymentService)
	log.Println("✓ MercadoPago client initialized")

	// Run pending migrations
	migrationRunner := migration.NewRunner(db.Database, migration.All())
	if err := migrationRunner.Run(context.Background()); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed plans after migrations (cleanCollections may have dropped the plans collection)
	if err := seedPlans(context.Background(), planRepository); err != nil {
		log.Printf("Warning: failed to seed plans: %v", err)
	}

	// Initialize Request module
	requestRepository := requestRepo.NewMongoRepository(db.Database)

	// Notification function for WebSocket
	notifyFunc := func(restaurantID primitive.ObjectID, request *requestDomain.Request) {
		hub.Broadcast(restaurantID, request)
	}

	requestService := requestUseCase.NewRequestUseCase(
		requestRepository,
		restaurantRepository,
		branchRepository,
		tableRepository,
		qrService,
		notifyFunc,
	)
	requestHdlr := requestHandler.NewRequestHandler(requestService, hub, jwtService)

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize Gin router
	router := gin.Default()

	// Sentry middleware (must be before other middleware)
	if cfg.SentryDSN != "" {
		router.Use(sentrygin.New(sentrygin.Options{
			Repanic: true,
		}))
	}

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	log.Printf("✓ CORS enabled for origins: %v", cfg.CORSAllowedOrigins)

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

			// Branch routes
			branchHdlr.RegisterRoutes(protected)

			// Table routes
			tableHdlr.RegisterRoutes(protected)

			// Setup routes
			setupHdlr.RegisterRoutes(protected)

			// Subscription routes (both public and protected)
			subscriptionHdlr.RegisterRoutes(protected, publicV1)

			// Payment routes (both public and protected)
			paymentHdlr.RegisterRoutes(protected, publicV1)

			// Request routes (both public and protected)
			requestHdlr.RegisterRoutes(protected, publicV1)

			// Invitation routes
			invitationHdlr.RegisterRoutes(protected)
		}

		// WebSocket route (no authentication middleware, but validates token in handler)
		requestHdlr.RegisterWebSocketRoute(v1)
	}

	log.Printf("🚀 Server running on port %s\n", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func seedPlans(ctx context.Context, repo subscriptionRepo.PlanRepository) error {
	existing, err := repo.FindAll(ctx)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}

	plans := []*subscriptionDomain.Plan{
		subscriptionDomain.NewPlan(subscriptionDomain.PlanNameBasico, 19.99, 20, 1, 30),
		subscriptionDomain.NewPlan(subscriptionDomain.PlanNameIntermedio, 49.99, 50, 3, 30),
		subscriptionDomain.NewPlan(subscriptionDomain.PlanNameProfesional, 199.99, subscriptionDomain.Unlimited, subscriptionDomain.Unlimited, 30),
	}

	for _, plan := range plans {
		if err := repo.Create(ctx, plan); err != nil {
			return err
		}
		log.Printf("✓ Plan '%s' sembrado con ID: %s", plan.Name, plan.ID.Hex())
	}

	return nil
}
