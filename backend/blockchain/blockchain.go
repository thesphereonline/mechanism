package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"0xygen.thesphere.online/backend/contracts"
)

var (
	client       *ethclient.Client
	sphereToken  *contracts.SphereToken
	sphereNFT    *contracts.SphereNFT
	adminKey     *ecdsa.PrivateKey
	adminAddress common.Address
	tokenAddress common.Address
	nftAddress   common.Address
	chainID      *big.Int
	gasLimit     uint64
	gasPrice     *big.Int
)

// InitBlockchain initializes the blockchain connection
func InitBlockchain() error {
	var err error

	// Connect to Ethereum node
	rpcURL := os.Getenv("ETHEREUM_RPC_URL")
	if rpcURL == "" {
		rpcURL = "http://localhost:8545" // Default to local node
	}

	client, err = ethclient.Dial(rpcURL)
	if err != nil {
		return fmt.Errorf("failed to connect to Ethereum node: %v", err)
	}

	// Get chain ID
	chainID, err = client.ChainID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %v", err)
	}

	// Set gas parameters
	gasLimitStr := os.Getenv("GAS_LIMIT")
	if gasLimitStr == "" {
		gasLimit = 3000000 // Default gas limit
	} else {
		gasLimitInt, err := strconv.ParseUint(gasLimitStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid gas limit: %v", err)
		}
		gasLimit = gasLimitInt
	}

	gasPriceStr := os.Getenv("GAS_PRICE")
	if gasPriceStr == "" {
		gasPrice = big.NewInt(20000000000) // 20 Gwei default
	} else {
		gasPrice, _ = new(big.Int).SetString(gasPriceStr, 10)
	}

	// Load admin private key
	adminKeyHex := os.Getenv("ADMIN_PRIVATE_KEY")
	if adminKeyHex == "" {
		return fmt.Errorf("admin private key not set")
	}

	adminKey, err = crypto.HexToECDSA(adminKeyHex)
	if err != nil {
		return fmt.Errorf("invalid admin private key: %v", err)
	}

	adminAddress = crypto.PubkeyToAddress(adminKey.PublicKey)

	// Load contract addresses
	tokenAddressHex := os.Getenv("TOKEN_CONTRACT_ADDRESS")
	if tokenAddressHex == "" {
		return fmt.Errorf("token contract address not set")
	}
	tokenAddress = common.HexToAddress(tokenAddressHex)

	nftAddressHex := os.Getenv("NFT_CONTRACT_ADDRESS")
	if nftAddressHex == "" {
		return fmt.Errorf("NFT contract address not set")
	}
	nftAddress = common.HexToAddress(nftAddressHex)

	// Initialize contract instances
	sphereToken, err = contracts.NewSphereToken(tokenAddress, client)
	if err != nil {
		return fmt.Errorf("failed to initialize token contract: %v", err)
	}

	sphereNFT, err = contracts.NewSphereNFT(nftAddress, client)
	if err != nil {
		return fmt.Errorf("failed to initialize NFT contract: %v", err)
	}

	log.Println("Blockchain connection initialized successfully")
	return nil
}

// MintNFT mints a new NFT
func MintNFT(recipient string, tokenURI string) (string, string, error) {
	// Create transaction options
	auth, err := createTransactionOpts()
	if err != nil {
		return "", "", err
	}

	// Mint NFT
	tx, err := sphereNFT.MintNFT(auth, common.HexToAddress(recipient), tokenURI)
	if err != nil {
		return "", "", fmt.Errorf("failed to mint NFT: %v", err)
	}

	// Wait for transaction to be mined
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		return "", "", fmt.Errorf("failed to wait for transaction: %v", err)
	}

	// Get token ID from logs
	tokenID := "0" // Default value
	if len(receipt.Logs) > 0 {
		// Parse logs to get token ID
		// This depends on the event structure in your contract
		// ...
	}

	return tokenID, tx.Hash().Hex(), nil
}

// ListNFT lists an NFT for sale
func ListNFT(owner string, tokenID string, price float64) (string, error) {
	// Create transaction options
	auth, err := createTransactionOpts()
	if err != nil {
		return "", err
	}

	// Convert price to wei (assuming price is in Sphere tokens)
	priceInWei := new(big.Int).Mul(
		big.NewInt(int64(price*100)),    // Convert to smallest unit (assuming 2 decimals)
		big.NewInt(1000000000000000000), // 10^18
	)
	priceInWei = priceInWei.Div(priceInWei, big.NewInt(100))

	// Convert token ID to big.Int
	tokenIDInt, ok := new(big.Int).SetString(tokenID, 10)
	if !ok {
		return "", fmt.Errorf("invalid token ID")
	}

	// List NFT
	tx, err := sphereNFT.ListNFT(auth, tokenIDInt, priceInWei)
	if err != nil {
		return "", fmt.Errorf("failed to list NFT: %v", err)
	}

	// Wait for transaction to be mined
	_, err = bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		return "", fmt.Errorf("failed to wait for transaction: %v", err)
	}

	return tx.Hash().Hex(), nil
}

// BuyNFT buys an NFT
func BuyNFT(buyer string, tokenID string) (string, error) {
	// Create transaction options
	auth, err := createTransactionOpts()
	if err != nil {
		return "", err
	}

	// Convert token ID to big.Int
	tokenIDInt, ok := new(big.Int).SetString(tokenID, 10)
	if !ok {
		return "", fmt.Errorf("invalid token ID")
	}

	// Buy NFT
	tx, err := sphereNFT.BuyNFT(auth, tokenIDInt)
	if err != nil {
		return "", fmt.Errorf("failed to buy NFT: %v", err)
	}

	// Wait for transaction to be mined
	_, err = bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		return "", fmt.Errorf("failed to wait for transaction: %v", err)
	}

	return tx.Hash().Hex(), nil
}

// Helper function to create transaction options
func createTransactionOpts() (*bind.TransactOpts, error) {
	nonce, err := client.PendingNonceAt(context.Background(), adminAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %v", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(adminKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %v", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice

	return auth, nil
}
