package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Block represents a single block in the chain
type Block struct {
	Index     int
	Timestamp string
	Data      string
	PrevHash  string
	Hash      string
	Nonce     int    // Used for PoW
	Validator string // dBFT Validator elected for this block
}

// calculateHash generates the SHA256 hash of a block
func calculateHash(block Block) string {
	record := fmt.Sprintf("%d%s%s%s%d%s", block.Index, block.Timestamp, block.Data, block.PrevHash, block.Nonce, block.Validator)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// createBlock creates a new block based on the previous one and adds a Proof of Work (PoW)
func createBlock(prevBlock Block, data string, validator string) Block {
	newBlock := Block{
		Index:     prevBlock.Index + 1,
		Timestamp: time.Now().String(),
		Data:      data,
		PrevHash:  prevBlock.Hash,
		Validator: validator,
	}

	// Run Proof of Work to find the valid nonce
	newBlock.Nonce = mineBlock(newBlock)
	newBlock.Hash = calculateHash(newBlock)

	// Simulate dBFT Consensus (Validator approves the block)
	if !dBFTConsensus(newBlock) {
		fmt.Println("Block not approved by dBFT!")
	}

	return newBlock
}

// createGenesisBlock returns the first block in the chain
func createGenesisBlock() Block {
	genesis := Block{
		Index:     0,
		Timestamp: time.Now().String(),
		Data:      "Genesis Block",
		PrevHash:  "",
	}
	genesis.Nonce = mineBlock(genesis) // No previous block for genesis
	genesis.Hash = calculateHash(genesis)
	return genesis
}

// mineBlock tries different nonces until it finds a valid hash
func mineBlock(block Block) int {
	// Difficulty level: how many leading zeros the hash should have
	const difficulty = 4
	var nonce int
	for {
		block.Nonce = nonce
		hash := calculateHash(block)
		if isValidHash(hash, difficulty) {
			return nonce
		}
		nonce++
	}
}

// isValidHash checks if the hash has enough leading zeros based on difficulty
func isValidHash(hash string, difficulty int) bool {
	prefix := ""
	for i := 0; i < difficulty; i++ {
		prefix += "0"
	}
	return hash[:difficulty] == prefix
}

func main() {
	// Create blockchain
	blockchain := []Block{createGenesisBlock()}

	// Add a few blocks with PoW and dBFT consensus
	blockchain = append(blockchain, createBlock(blockchain[len(blockchain)-1], "Block 1 Data", "Validator1"))
	blockchain = append(blockchain, createBlock(blockchain[len(blockchain)-1], "Block 2 Data", "Validator2"))

	// Print chain
	for _, block := range blockchain {
		fmt.Printf("Index: %d\nTimestamp: %s\nData: %s\nPrevHash: %s\nHash: %s\nNonce: %d\nValidator: %s\n\n",
			block.Index, block.Timestamp, block.Data, block.PrevHash, block.Hash, block.Nonce, block.Validator)
	}
}
