package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"0xygen.thesphere.online/backend/blockchain"
	"0xygen.thesphere.online/backend/database"
	"0xygen.thesphere.online/backend/models"
	"0xygen.thesphere.online/backend/storage"
)

// GetAllNFTs returns all NFTs in the marketplace
func GetAllNFTs(c *gin.Context) {
	var nfts []models.NFT

	// Get query parameters for pagination
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "20")

	// Get filter parameters
	category := c.Query("category")
	minPrice := c.Query("minPrice")
	maxPrice := c.Query("maxPrice")

	// Build query
	query := database.DB.Model(&models.NFT{})

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if minPrice != "" {
		query = query.Where("price >= ?", minPrice)
	}

	if maxPrice != "" {
		query = query.Where("price <= ?", maxPrice)
	}

	// Execute query with pagination
	result := query.Scopes(database.Paginate(page, limit)).Find(&nfts)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch NFTs"})
		return
	}

	c.JSON(http.StatusOK, nfts)
}

// GetNFTByID returns a specific NFT by ID
func GetNFTByID(c *gin.Context) {
	id := c.Param("id")

	var nft models.NFT
	result := database.DB.First(&nft, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "NFT not found"})
		return
	}

	c.JSON(http.StatusOK, nft)
}

// UploadNFT handles the upload of NFT artwork
func UploadNFT(c *gin.Context) {
	// Get user from context (set by auth middleware)
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse multipart form
	err := c.Request.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	// Get form data
	title := c.PostForm("title")
	description := c.PostForm("description")
	category := c.PostForm("category")

	if title == "" || description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title and description are required"})
		return
	}

	// Get file
	file, header, err := c.Request.FormFile("artwork")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Check file type
	ext := filepath.Ext(header.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only image files are allowed"})
		return
	}

	// Generate unique filename
	filename := uuid.New().String() + ext

	// Upload to storage
	fileURL, err := storage.UploadFile(file, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
		return
	}

	// Create NFT metadata
	metadata := map[string]interface{}{
		"name":        title,
		"description": description,
		"image":       fileURL,
		"creator":     user.(models.User).Address,
		"created_at":  time.Now().Unix(),
		"attributes": []map[string]string{
			{"trait_type": "Category", "value": category},
		},
	}

	// Upload metadata to IPFS
	metadataURL, err := storage.UploadMetadata(metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload metadata"})
		return
	}

	// Create NFT record in database
	nft := models.NFT{
		Title:       title,
		Description: description,
		Category:    category,
		ImageURL:    fileURL,
		MetadataURL: metadataURL,
		CreatorID:   user.(models.User).ID,
		Status:      "uploaded", // Not yet minted
	}

	result := database.DB.Create(&nft)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save NFT"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "NFT uploaded successfully",
		"nft":     nft,
	})
}

// MintNFT mints an uploaded NFT on the blockchain
func MintNFT(c *gin.Context) {
	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse request
	var req struct {
		NFTID uint `json:"nft_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get NFT from database
	var nft models.NFT
	result := database.DB.First(&nft, req.NFTID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "NFT not found"})
		return
	}

	// Check if user is the creator
	if nft.CreatorID != user.(models.User).ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the creator can mint this NFT"})
		return
	}

	// Check if NFT is already minted
	if nft.Status != "uploaded" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "NFT is already minted or listed"})
		return
	}

	// Mint NFT on blockchain
	tokenID, txHash, err := blockchain.MintNFT(user.(models.User).Address, nft.MetadataURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to mint NFT: %v", err)})
		return
	}

	// Update NFT in database
	nft.TokenID = tokenID
	nft.TxHash = txHash
	nft.Status = "minted"

	result = database.DB.Save(&nft)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update NFT status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "NFT minted successfully",
		"nft":     nft,
	})
}

// ListNFT lists an NFT for sale
func ListNFT(c *gin.Context) {
	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse request
	var req struct {
		NFTID uint    `json:"nft_id" binding:"required"`
		Price float64 `json:"price" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get NFT from database
	var nft models.NFT
	result := database.DB.First(&nft, req.NFTID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "NFT not found"})
		return
	}

	// Check if user is the owner
	if nft.OwnerID != user.(models.User).ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the owner can list this NFT"})
		return
	}

	// Check if NFT is minted
	if nft.Status != "minted" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "NFT must be minted before listing"})
		return
	}

	// List NFT on blockchain
	txHash, err := blockchain.ListNFT(user.(models.User).Address, nft.TokenID, req.Price)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to list NFT: %v", err)})
		return
	}

	// Update NFT in database
	nft.Price = req.Price
	nft.ListingTxHash = txHash
	nft.Status = "listed"

	result = database.DB.Save(&nft)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update NFT status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "NFT listed successfully",
		"nft":     nft,
	})
}

// BuyNFT handles the purchase of an NFT
func BuyNFT(c *gin.Context) {
	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse request
	var req struct {
		NFTID uint `json:"nft_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get NFT from database
	var nft models.NFT
	result := database.DB.First(&nft, req.NFTID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "NFT not found"})
		return
	}

	// Check if NFT is listed
	if nft.Status != "listed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "NFT is not listed for sale"})
		return
	}

	// Check if user is not the owner
	if nft.OwnerID == user.(models.User).ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot buy your own NFT"})
		return
	}

	// Buy NFT on blockchain
	txHash, err := blockchain.BuyNFT(user.(models.User).Address, nft.TokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to buy NFT: %v", err)})
		return
	}

	// Update NFT in database
	previousOwnerID := nft.OwnerID
	nft.OwnerID = user.(models.User).ID
	nft.SaleTxHash = txHash
	nft.Status = "owned"

	result = database.DB.Save(&nft)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update NFT ownership"})
		return
	}

	// Record transaction
	transaction := models.Transaction{
		Type:      "nft_purchase",
		FromID:    previousOwnerID,
		ToID:      user.(models.User).ID,
		NFTID:     nft.ID,
		Amount:    nft.Price,
		TxHash:    txHash,
		Timestamp: time.Now(),
	}

	result = database.DB.Create(&transaction)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "NFT purchased successfully",
		"nft":     nft,
	})
}

// GetUserNFTs returns all NFTs owned by the user
func GetUserNFTs(c *gin.Context) {
	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var nfts []models.NFT
	result := database.DB.Where("owner_id = ?", user.(models.User).ID).Find(&nfts)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch NFTs"})
		return
	}

	c.JSON(http.StatusOK, nfts)
}
