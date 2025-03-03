package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// GetTokenPrice gets the current token price from the blockchain
func GetTokenPrice() (float64, error) {
	// Get token price in wei
	priceInWei, err := sphereToken.TokenPriceInWei(nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get token price: %v", err)
	}

	// Convert wei to ETH (1 ETH = 10^18 wei)
	priceInEth := new(big.Float).Quo(
		new(big.Float).SetInt(priceInWei),
		new(big.Float).SetInt(big.NewInt(1000000000000000000)),
	)

	// Convert to float64
	price, _ := priceInEth.Float64()
	return price, nil
}

// UpdateTokenPrice updates the token price on the blockchain
func UpdateTokenPrice(priceInEth float64) error {
	// Create transaction options
	auth, err := createTransactionOpts()
	if err != nil {
		return err
	}

	// Convert ETH to wei (1 ETH = 10^18 wei)
	priceInWei := new(big.Int).Mul(
		big.NewInt(int64(priceInEth*1000000)),
		big.NewInt(1000000000000),
	)

	// Update token price
	tx, err := sphereToken.UpdateTokenPrice(auth, priceInWei)
	if err != nil {
		return fmt.Errorf("failed to update token price: %v", err)
	}

	// Wait for transaction to be mined
	_, err = bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction: %v", err)
	}

	return nil
}

// MintTokens mints new tokens to a user
func MintTokens(recipient string, amount float64) (string, error) {
	// Create transaction options
	auth, err := createTransactionOpts()
	if err != nil {
		return "", err
	}

	// Mint tokens
	tx, err := sphereToken.MintTokens(auth, common.HexToAddress(recipient), big.NewInt(int64(amount)))
	if err != nil {
		return "", fmt.Errorf("failed to mint tokens: %v", err)
	}

	// Wait for transaction to be mined
	_, err = bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		return "", fmt.Errorf("failed to wait for transaction: %v", err)
	}

	return tx.Hash().Hex(), nil
}

// RecordFiatPurchase records a fiat purchase on the blockchain
func RecordFiatPurchase(buyer string, amount float64, referenceID string) (string, error) {
	// Create transaction options
	auth, err := createTransactionOpts()
	if err != nil {
		return "", err
	}

	// Record fiat purchase
	tx, err := sphereToken.RecordFiatPurchase(
		auth,
		common.HexToAddress(buyer),
		big.NewInt(int64(amount)),
		referenceID,
	)
	if err != nil {
		return "", fmt.Errorf("failed to record fiat purchase: %v", err)
	}

	// Wait for transaction to be mined
	_, err = bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		return "", fmt.Errorf("failed to wait for transaction: %v", err)
	}

	return tx.Hash().Hex(), nil
}
