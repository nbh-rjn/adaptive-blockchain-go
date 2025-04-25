package main

import (
	"fmt"
	"time"
)

// Block represents a single block in a shard
type Block struct {
	Index     int
	Timestamp string
	Data      string
	PrevHash  string
	Hash      string
	Nonce     int
	Validator string
}

// Genesis block for a shard
func createGenesisBlock() Block {
	genesis := Block{
		Index:     0,
		Timestamp: time.Now().String(),
		Data:      "Genesis Block",
		PrevHash:  "",
	}
	genesis.Nonce = mineBlock(genesis)
	genesis.Hash = calculateHash(genesis)
	return genesis
}

func main() {
	initAMQFilters()

	// Initialize shards with genesis blocks
	for i := 0; i < shardCount; i++ {
		genesis := createGenesisBlock()
		merkleForest = append(merkleForest, Shard{
			Blocks:     []Block{genesis},
			MerkleRoot: genesis.Hash,
		})
	}

	// Add some blocks
	addBlockToShards("Block A", "Validator1")
	addBlockToShards("Block B", "Validator2")
	addBlockToShards("Block C", "Validator1")
	addBlockToShards("Block D", "Validator2")

	// Example of interacting with CAP orchestration
	// You can dynamically switch the state to simulate different network conditions.
	// This can be tied to actual network conditions or user commands.

	// For demonstration, we will switch to different CAP modes:
	fmt.Println("Starting CAP Orchestration...")
	CAPOrchestrator()

	proof := generateMerkleProof(0, 2)
	fmt.Println("Merkle Proof:", proof)

	// Print each shard and its Merkle root
	for i, shard := range merkleForest {
		fmt.Printf("Shard %d (Merkle Root: %s)\n", i, shard.MerkleRoot)
		for _, block := range shard.Blocks {
			fmt.Printf("  Block %d: %s\n", block.Index, block.Hash[:10])
		}
		fmt.Println()
	}

	// Synchronize shards to update Merkle roots
	synchronizeShards()

	// Check AMQ presence
	hash := merkleForest[0].Blocks[0].Hash
	fmt.Println("Is genesis in AMQ of Shard 0?", isInAMQ(0, hash))

	// Show compressed Merkle proof
	compressed := compressMerkleProof(proof)
	fmt.Println("Compressed Merkle Proof:", compressed)

	// Show accumulator snapshot
	snapshot := getAccumulatorSnapshot(0)
	fmt.Println("Accumulator Snapshot (Shard 0):", snapshot)
	// Simulate vector clock updates
	applyVectorClocks()

	// Conflict resolution simulation
	resolveConflicts()
}
