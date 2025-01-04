package main

import (
	"fmt"
	"sync"
)

// Transaction represents a simple blockchain transaction.
type Transaction struct {
	ID       int
	Amount   float64
	Sender   string
	Receiver string
}

// Block represents a block in the blockchain.
type Block struct {
	Index        int
	Transactions []Transaction
	PreviousHash string
	Hash         string
}

// Node represents a blockchain peer node.
type Node struct {
	ID         string
	Blockchain []Block
	Mutex      sync.Mutex
}

// CreateGenesisBlock initializes the blockchain with the genesis block.
func CreateGenesisBlock() Block {
	return Block{
		Index:        0,
		Transactions: []Transaction{},
		PreviousHash: "0",
		Hash:         "genesis_hash",
	}
}

// AddTransaction adds a transaction to the latest block.
func (node *Node) AddTransaction(tx Transaction) {
	node.Mutex.Lock()
	defer node.Mutex.Unlock()

	latestBlock := &node.Blockchain[len(node.Blockchain)-1]
	latestBlock.Transactions = append(latestBlock.Transactions, tx)
	fmt.Printf("Transaction added to Node %s: %+v\n", node.ID, tx)
}

// AddBlock adds a new block to the blockchain.
func (node *Node) AddBlock(newBlock Block) {
	node.Mutex.Lock()
	defer node.Mutex.Unlock()

	node.Blockchain = append(node.Blockchain, newBlock)
	fmt.Printf("Block added to Node %s: %+v\n", node.ID, newBlock)
}

// SimulateNetwork demonstrates the network of interconnected nodes.
func SimulateNetwork() {
	// Create nodes
	nodeA := Node{ID: "A", Blockchain: []Block{CreateGenesisBlock()}}
	nodeB := Node{ID: "B", Blockchain: []Block{CreateGenesisBlock()}}
	nodeC := Node{ID: "C", Blockchain: []Block{CreateGenesisBlock()}}

	// Simulate adding transactions
	tx1 := Transaction{ID: 1, Amount: 100, Sender: "Alice", Receiver: "Bob"}
	nodeA.AddTransaction(tx1)
	nodeB.AddTransaction(tx1)
	nodeC.AddTransaction(tx1)

	// Simulate adding a new block
	newBlock := Block{
		Index:        1,
		Transactions: []Transaction{tx1},
		PreviousHash: "genesis_hash",
		Hash:         "block_1_hash",
	}
	nodeA.AddBlock(newBlock)
	nodeB.AddBlock(newBlock)
	nodeC.AddBlock(newBlock)

	// Print blockchains
	fmt.Printf("Node A Blockchain: %+v\n", nodeA.Blockchain)
	fmt.Printf("Node B Blockchain: %+v\n", nodeB.Blockchain)
	fmt.Printf("Node C Blockchain: %+v\n", nodeC.Blockchain)
}

func main() {
	SimulateNetwork()
}
