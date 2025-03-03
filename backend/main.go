package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"0xygen.thesphere.online/backend/controllers"
	"0xygen.thesphere.online/backend/database"
	"0xygen.thesphere.online/backend/middleware"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using environment variables")
	}

	// Initialize database
	database.InitDB()

	// Set up Gin router
	router := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000", "https://0xygen.thesphere.online"}
	config.AllowCredentials = true
	config.AddAllowHeaders("Authorization")
	router.Use(cors.New(config))

	// API routes
	api := router.Group("/api")
	{
		// Public routes
		api.GET("/nfts", controllers.GetAllNFTs)
		api.GET("/nfts/:id", controllers.GetNFTByID)
		api.GET("/token/price", controllers.GetTokenPrice)

		// Protected routes
		authorized := api.Group("/")
		authorized.Use(middleware.AuthMiddleware())
		{
			// NFT routes
			authorized.POST("/nfts/upload", controllers.UploadNFT)
			authorized.POST("/nfts/mint", controllers.MintNFT)
			authorized.POST("/nfts/list", controllers.ListNFT)
			authorized.POST("/nfts/buy", controllers.BuyNFT)

			// Token routes
			authorized.POST("/token/buy", controllers.BuyTokenWithFiat)

			// User routes
			authorized.GET("/user/nfts", controllers.GetUserNFTs)
			authorized.GET("/user/transactions", controllers.GetUserTransactions)
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(middleware.AdminMiddleware())
		{
			admin.POST("/token/update-price", controllers.UpdateTokenPrice)
			admin.GET("/transactions", controllers.GetAllTransactions)
			admin.POST("/fiat/confirm", controllers.ConfirmFiatPayment)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router.Run(":" + port)
}
