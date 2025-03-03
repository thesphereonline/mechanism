package core

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// Block represents a single block in the blockchain
type Block struct {
	Index        int64
	Timestamp    int64
	Transactions []Transaction
	PrevHash     string
	Hash         string
	Nonce        int64
}

// Transaction represents a transaction on the blockchain
type Transaction struct {
	ID        string
	From      string
	To        string
	Amount    float64
	Timestamp int64
	Signature string
	Data      map[string]interface{} // For NFT metadata
}

// Blockchain represents the entire blockchain
type Blockchain struct {
	Chain               []*Block
	PendingTransactions []Transaction
	Difficulty          int
	MiningReward        float64
	Nodes               []string
}

// NewBlockchain creates a new blockchain with a genesis block
func NewBlockchain(difficulty int, miningReward float64) *Blockchain {
	blockchain := &Blockchain{
		Chain:               []*Block{},
		PendingTransactions: []Transaction{},
		Difficulty:          difficulty,
		MiningReward:        miningReward,
		Nodes:               []string{},
	}

	// Create genesis block
	genesisBlock := &Block{
		Index:        0,
		Timestamp:    time.Now().Unix(),
		Transactions: []Transaction{},
		PrevHash:     "0",
		Nonce:        0,
	}
	genesisBlock.Hash = calculateHash(genesisBlock)
	blockchain.Chain = append(blockchain.Chain, genesisBlock)

	return blockchain
}

// calculateHash calculates the hash of a block
func calculateHash(block *Block) string {
	record := string(block.Index) + string(block.Timestamp) + block.PrevHash + string(block.Nonce)
	for _, tx := range block.Transactions {
		record += tx.ID
	}

	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// AddTransaction adds a new transaction to pending transactions
func (bc *Blockchain) AddTransaction(tx Transaction) bool {
	// Verify transaction signature here
	// ...

	bc.PendingTransactions = append(bc.PendingTransactions, tx)
	return true
}

// MinePendingTransactions mines pending transactions into a new block
func (bc *Blockchain) MinePendingTransactions(minerAddress string) {
	// Create mining reward transaction
	rewardTx := Transaction{
		ID:        generateTransactionID(),
		From:      "SYSTEM",
		To:        minerAddress,
		Amount:    bc.MiningReward,
		Timestamp: time.Now().Unix(),
		Data:      map[string]interface{}{"type": "mining_reward"},
	}

	bc.PendingTransactions = append(bc.PendingTransactions, rewardTx)

	// Create new block
	block := &Block{
		Index:        int64(len(bc.Chain)),
		Timestamp:    time.Now().Unix(),
		Transactions: bc.PendingTransactions,
		PrevHash:     bc.Chain[len(bc.Chain)-1].Hash,
		Nonce:        0,
	}

	// Mine the block (proof of work)
	bc.mineBlock(block)

	// Add block to chain
	bc.Chain = append(bc.Chain, block)

	// Clear pending transactions
	bc.PendingTransactions = []Transaction{}
}

// mineBlock mines a block (proof of work)
func (bc *Blockchain) mineBlock(block *Block) {
	target := string(make([]byte, bc.Difficulty))

	for {
		block.Hash = calculateHash(block)
		if block.Hash[:bc.Difficulty] == target {
			break
		}
		block.Nonce++
	}
}

// IsChainValid checks if the blockchain is valid
func (bc *Blockchain) IsChainValid() bool {
	for i := 1; i < len(bc.Chain); i++ {
		currentBlock := bc.Chain[i]
		prevBlock := bc.Chain[i-1]

		// Check if hash is correct
		if currentBlock.Hash != calculateHash(currentBlock) {
			return false
		}

		// Check if previous hash is correct
		if currentBlock.PrevHash != prevBlock.Hash {
			return false
		}
	}

	return true
}

// Helper function to generate transaction ID
func generateTransactionID() string {
	// Generate a unique ID based on timestamp and random number
	// ...
	return "tx_" + hex.EncodeToString([]byte(time.Now().String()))
}
