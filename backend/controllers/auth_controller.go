package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"

	"0xygen.thesphere.online/backend/database"
	"0xygen.thesphere.online/backend/models"
)

// LoginRequest represents a login request
type LoginRequest struct {
	Address   string `json:"address" binding:"required"`
	Signature string `json:"signature" binding:"required"`
	Nonce     string `json:"nonce" binding:"required"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Address  string `json:"address" binding:"required"`
	Username string `json:"username" binding:"required"`
	Email    string `json:"email"`
}

// GetNonce generates a nonce for authentication
func GetNonce(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Address is required"})
		return
	}

	// Validate Ethereum address
	if !common.IsHexAddress(address) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Ethereum address"})
		return
	}

	// Generate random nonce
	nonceBytes := make([]byte, 32)
	_, err := rand.Read(nonceBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate nonce"})
		return
	}
	nonce := hex.EncodeToString(nonceBytes)

	// Check if user exists
	var user models.User
	result := database.DB.Where("address = ?", address).First(&user)
	if result.Error != nil {
		// User doesn't exist, return nonce for registration
		c.JSON(http.StatusOK, gin.H{
			"nonce":   nonce,
			"message": "Please sign this message to verify your ownership of this address: " + nonce,
			"exists":  false,
		})
		return
	}

	// User exists, return nonce for login
	c.JSON(http.StatusOK, gin.H{
		"nonce":    nonce,
		"message":  "Please sign this message to login: " + nonce,
		"exists":   true,
		"username": user.Username,
	})
}

// Register registers a new user
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate Ethereum address
	if !common.IsHexAddress(req.Address) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Ethereum address"})
		return
	}

	// Check if user already exists
	var existingUser models.User
	result := database.DB.Where("address = ?", req.Address).First(&existingUser)
	if result.Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User with this address already exists"})
		return
	}

	// Create new user
	user := models.User{
		Address:   req.Address,
		Username:  req.Username,
		Email:     req.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result = database.DB.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token, err := generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"address":  user.Address,
			"username": user.Username,
			"email":    user.Email,
			"is_admin": user.IsAdmin,
		},
	})
}

// Login authenticates a user
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate Ethereum address
	if !common.IsHexAddress(req.Address) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Ethereum address"})
		return
	}

	// Verify signature
	message := "Please sign this message to login: " + req.Nonce
	verified := verifySignature(req.Address, message, req.Signature)
	if !verified {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Get user from database
	var user models.User
	result := database.DB.Where("address = ?", req.Address).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Generate JWT token
	token, err := generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"address":  user.Address,
			"username": user.Username,
			"email":    user.Email,
			"is_admin": user.IsAdmin,
		},
	})
}

// Helper function to verify signature
func verifySignature(address, message, signature string) bool {
	// Convert signature to bytes
	sig, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	// Add Ethereum message prefix
	prefixedMessage := "\x19Ethereum Signed Message:\n" + string(len(message)) + message

	// Hash the message
	messageHash := crypto.Keccak256Hash([]byte(prefixedMessage))

	// Recover public key from signature
	if sig[64] >= 27 {
		sig[64] -= 27
	}

	pubKey, err := crypto.Ecrecover(messageHash.Bytes(), sig)
	if err != nil {
		return false
	}

	// Convert public key to address
	recoveredAddress := common.BytesToAddress(crypto.Keccak256(pubKey[1:])[12:])

	// Compare addresses
	return recoveredAddress.Hex() == address
}

// Helper function to generate JWT token
func generateToken(user models.User) (string, error) {
	// Set token expiration time
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create claims
	claims := jwt.MapClaims{
		"id":       user.ID,
		"address":  user.Address,
		"username": user.Username,
		"is_admin": user.IsAdmin,
		"exp":      expirationTime.Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
