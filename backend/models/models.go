package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Address   string         `json:"address" gorm:"unique;not null"`
	Username  string         `json:"username"`
	Email     string         `json:"email"`
	IsAdmin   bool           `json:"is_admin" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// NFT represents an NFT in the marketplace
type NFT struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	Title         string         `json:"title" gorm:"not null"`
	Description   string         `json:"description" gorm:"type:text"`
	Category      string         `json:"category"`
	ImageURL      string         `json:"image_url" gorm:"not null"`
	MetadataURL   string         `json:"metadata_url" gorm:"not null"`
	TokenID       string         `json:"token_id"`
	Price         float64        `json:"price" gorm:"default:0"`
	CreatorID     uint           `json:"creator_id" gorm:"not null"`
	Creator       User           `json:"creator" gorm:"foreignKey:CreatorID"`
	OwnerID       uint           `json:"owner_id"`
	Owner         User           `json:"owner" gorm:"foreignKey:OwnerID"`
	Status        string         `json:"status" gorm:"default:'uploaded'"`
	TxHash        string         `json:"tx_hash"`
	ListingTxHash string         `json:"listing_tx_hash"`
	SaleTxHash    string         `json:"sale_tx_hash"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// Transaction represents a transaction in the marketplace
type Transaction struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Type      string         `json:"type" gorm:"not null"` // token_purchase, nft_purchase, etc.
	FromID    uint           `json:"from_id"`
	From      User           `json:"from" gorm:"foreignKey:FromID"`
	ToID      uint           `json:"to_id"`
	To        User           `json:"to" gorm:"foreignKey:ToID"`
	NFTID     uint           `json:"nft_id"`
	NFT       NFT            `json:"nft" gorm:"foreignKey:NFTID"`
	Amount    float64        `json:"amount" gorm:"not null"`
	TxHash    string         `json:"tx_hash" gorm:"not null"`
	Status    string         `json:"status" gorm:"default:'completed'"`
	Timestamp time.Time      `json:"timestamp"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
