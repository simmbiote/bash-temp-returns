package main

import (
	"log"
	"os"

	"customer-support-api/internal/application/usecases"
	"customer-support-api/internal/domain/services"
	"customer-support-api/internal/infrastructure/clients/http"
	"customer-support-api/internal/infrastructure/clients/mock"
	"customer-support-api/internal/infrastructure/clients/openai"
	"customer-support-api/internal/infrastructure/persistence/sqlite"
	"customer-support-api/internal/interfaces/http/handlers"

	_ "customer-support-api/docs" // Import generated docs

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Customer Support API
// @version 0.2.0
// @description API for customer support conversations and return request management with Orders API integration
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@bash.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/chatbot.db"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	db, err := sqlite.NewDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	conversationRepo := sqlite.NewConversationRepository(db)
	messageRepo := sqlite.NewMessageRepository(db)
	returnRequestRepo := sqlite.NewReturnRequestRepository(db)

	// Initialize Orders API client
	ordersAPIURL := os.Getenv("ORDERS_API_URL")
	var ordersAPIClient services.OrdersAPIClient

	if ordersAPIURL == "" || ordersAPIURL == "mock" {
		log.Println("Using mock Orders API client")
		ordersAPIClient = mock.NewMockOrdersAPIClient()
	} else {
		log.Printf("Using HTTP Orders API client: %s", ordersAPIURL)
		ordersAPIClient = http.NewOrdersAPIClient(ordersAPIURL)
	}

	// Initialize AI client
	aiProvider := os.Getenv("AI_PROVIDER")
	var aiClient services.AIClient

	if aiProvider == "" || aiProvider == "mock" {
		log.Println("Using mock AI client")
		aiClient = mock.NewMockAIClient()
	} else if aiProvider == "openai" {
		openaiAPIKey := os.Getenv("OPENAI_API_KEY")
		openaiModel := os.Getenv("OPENAI_MODEL")
		if openaiModel == "" {
			openaiModel = "gpt-4o" // Default model
		}
		log.Printf("Using OpenAI client with model: %s", openaiModel)
		aiClient = openai.NewOpenAIClient(openaiAPIKey, openaiModel)
	} else {
		log.Fatalf("Unknown AI_PROVIDER: %s (supported: mock, openai)", aiProvider)
	}

	// Initialize use cases
	createConvUseCase := usecases.NewCreateConversationUseCase(conversationRepo, messageRepo)
	createReturnUseCase := usecases.NewCreateReturnRequestUseCase(returnRequestRepo, conversationRepo, ordersAPIClient)

	// Initialize additional use cases for CRUD operations
	getConversationUseCase := usecases.NewGetConversationUseCase(conversationRepo, messageRepo)
	sendMessageUseCase := usecases.NewSendMessageUseCase(conversationRepo, messageRepo, aiClient)
	getReturnUseCase := usecases.NewGetReturnRequestUseCase(returnRequestRepo)
	listReturnsUseCase := usecases.NewListReturnRequestsUseCase(returnRequestRepo)

	// Initialize handlers
	conversationHandler := handlers.NewConversationHandler(createConvUseCase, getConversationUseCase, sendMessageUseCase)
	returnHandler := handlers.NewReturnHandler(createReturnUseCase, getReturnUseCase, listReturnsUseCase)

	// Initialize Gin router
	router := gin.Default()

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		// Test database connection
		dbStatus := "connected"
		if err := db.Ping(); err != nil {
			dbStatus = "disconnected"
		}

		c.JSON(200, gin.H{
			"status":    "ok",
			"database":  os.Getenv("DB_TYPE"),
			"db_status": dbStatus,
			"cache":     os.Getenv("CACHE_TYPE"),
			"version":   "0.1.0",
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Conversation routes
		conversations := api.Group("/conversations")
		{
			conversations.POST("", conversationHandler.CreateConversation)
			conversations.GET("/:id", conversationHandler.GetConversation)
			conversations.GET("/:id/messages", conversationHandler.GetMessages)
			conversations.POST("/:id/messages", conversationHandler.SendMessage)
		}

		// Return routes
		returns := api.Group("/returns")
		{
			returns.POST("", returnHandler.CreateReturn)
			returns.GET("/:id", returnHandler.GetReturn)
			returns.GET("", returnHandler.ListReturns)
		}

		// Ping endpoint for testing
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message":   "pong",
				"timestamp": "2026-02-04T12:00:00Z",
			})
		})
	}

	// Start server
	log.Printf("🚀 Starting Customer Support API on port %s...", port)
	log.Printf("📦 Database: %s (%s)", os.Getenv("DB_TYPE"), dbPath)
	log.Printf("💾 Cache: %s", os.Getenv("CACHE_TYPE"))
	log.Printf("🔧 Log Level: %s", logLevel)
	log.Printf("\n✨ API endpoints:")
	log.Printf("   GET  /health")
	log.Printf("   GET  /api/v1/ping")
	log.Printf("   POST /api/v1/conversations")
	log.Printf("   POST /api/v1/returns")

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
