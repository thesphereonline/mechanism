package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"0xygen.thesphere.online/backend/blockchain"
	"0xygen.thesphere.online/backend/database"
	"0xygen.thesphere.online/backend/models"
)

// GetTokenPrice returns the current token price
func GetTokenPrice(c *gin.Context) {
	price, err := blockchain.GetTokenPrice()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get token price"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"price": price})
}

// BuyTokenWithFiat initiates a token purchase with fiat
func BuyTokenWithFiat(c *gin.Context) {
	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse request
	var req struct {
		Amount      float64 `json:"amount" binding:"required"`
		PaymentType string  `json:"payment_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate amount
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount must be greater than zero"})
		return
	}

	// Validate payment type
	if req.PaymentType != "credit_card" && req.PaymentType != "bank_transfer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment type"})
		return
	}

	// Generate reference ID
	referenceID := uuid.New().String()

	// Record transaction in database
	transaction := models.Transaction{
		Type:      "token_purchase_fiat",
		ToID:      user.(models.User).ID,
		Amount:    req.Amount,
		TxHash:    referenceID,
		Status:    "pending",
		Timestamp: time.Now(),
	}

	result := database.DB.Create(&transaction)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record transaction"})
		return
	}

	// Return payment instructions
	c.JSON(http.StatusOK, gin.H{
		"message":      "Token purchase initiated",
		"reference_id": referenceID,
		"instructions": "Please complete the payment using the provided reference ID",
	})
}

// UpdateTokenPrice updates the token price (admin only)
func UpdateTokenPrice(c *gin.Context) {
	// Parse request
	var req struct {
		Price float64 `json:"price" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate price
	if req.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price must be greater than zero"})
		return
	}

	// Update price on blockchain
	err := blockchain.UpdateTokenPrice(req.Price)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update token price"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Token price updated successfully"})
}

// ConfirmFiatPayment confirms a fiat payment (admin only)
func ConfirmFiatPayment(c *gin.Context) {
	// Parse request
	var req struct {
		ReferenceID string `json:"reference_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get transaction from database
	var transaction models.Transaction
	result := database.DB.First(&transaction, "tx_hash = ? AND type = 'token_purchase_fiat' AND status = 'pending'", req.ReferenceID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	// Get user
	var user models.User
	result = database.DB.First(&user, transaction.ToID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// Mint tokens on blockchain
	txHash, err := blockchain.MintTokens(user.Address, transaction.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mint tokens"})
		return
	}

	// Update transaction in database
	transaction.Status = "completed"
	transaction.TxHash = txHash
	result = database.DB.Save(&transaction)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment confirmed and tokens minted successfully"})
}

// GetUserTransactions returns all transactions for the user
func GetUserTransactions(c *gin.Context) {
	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var transactions []models.Transaction
	result := database.DB.Where("from_id = ? OR to_id = ?", user.(models.User).ID, user.(models.User).ID).
		Order("timestamp DESC").
		Find(&transactions)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

// GetAllTransactions returns all transactions (admin only)
func GetAllTransactions(c *gin.Context) {
	var transactions []models.Transaction
	result := database.DB.Order("timestamp DESC").Find(&transactions)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}
