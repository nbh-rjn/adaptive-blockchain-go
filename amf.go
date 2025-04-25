package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Shard represents a mini-blockchain (shard) with blocks and a Merkle root
type Shard struct {
	Blocks     []Block
	MerkleRoot string
}

// Global Merkle Forest (list of shards)
var merkleForest []Shard

const (
	shardCount       = 2 // number of shards
	maxShardCapacity = 5 // maximum blocks in a shard before rebalancing
)

// Adds a block to the shard with fewest blocks (simple adaptive sharding)
func addBlockToShards(data string, validator string) {
	// Find shard with fewest blocks
	target := 0
	for i := 1; i < len(merkleForest); i++ {
		if len(merkleForest[i].Blocks) < len(merkleForest[target].Blocks) {
			target = i
		}
	}

	// Add block to selected shard
	shard := &merkleForest[target]
	prevBlock := shard.Blocks[len(shard.Blocks)-1]
	newBlock := Block{
		Index:     prevBlock.Index + 1,
		Timestamp: time.Now().String(),
		Data:      data,
		PrevHash:  prevBlock.Hash,
		Validator: validator,
	}
	newBlock.Nonce = mineBlock(newBlock)
	newBlock.Hash = calculateHash(newBlock)

	// Add block if dBFT approves
	if dBFTConsensus(newBlock) {
		shard.Blocks = append(shard.Blocks, newBlock)
		shard.MerkleRoot = updateMerkleRoot(shard.Blocks)

		// Check if shard size exceeds limit and rebalance if necessary
		if len(shard.Blocks) > maxShardCapacity {
			rebalanceShards()
		}

		// Synchronize state across shards (if needed)
		// For example, synchronizing with the next shard (example case)
		// You could sync with another specific shard based on your logic
		synchronizeStateAcrossShards(target, (target+1)%len(merkleForest)) // Sync with the next shard (circularly)
	} else {
		fmt.Println("Block rejected by dBFT.")
	}
}

// Simple Merkle root calculation (concatenated block hashes hashed together)
func updateMerkleRoot(blocks []Block) string {
	if len(blocks) == 0 {
		return ""
	}
	var combinedHashes string
	for _, block := range blocks {
		combinedHashes += block.Hash
	}
	sum := sha256.Sum256([]byte(combinedHashes))
	return hex.EncodeToString(sum[:])
}

// Generate Merkle Proof for a block in a shard (proof of inclusion)
func generateMerkleProof(shardIndex, blockIndex int) []string {
	shard := &merkleForest[shardIndex]
	proof := []string{}

	// Traverse the shard's blocks and generate proof
	for i := blockIndex + 1; i < len(shard.Blocks); i++ {
		proof = append(proof, shard.Blocks[i].Hash)
	}

	return proof
}

// Rebalance the shards by moving blocks from the larger shards to smaller ones
func rebalanceShards() {
	// Find the shard with the most blocks
	var maxShardIndex int
	maxBlockCount := 0
	for i, shard := range merkleForest {
		if len(shard.Blocks) > maxBlockCount {
			maxShardIndex = i
			maxBlockCount = len(shard.Blocks)
		}
	}

	// Find the shard with the fewest blocks
	var minShardIndex int
	minBlockCount := len(merkleForest[0].Blocks)
	for i, shard := range merkleForest {
		if len(shard.Blocks) < minBlockCount {
			minShardIndex = i
			minBlockCount = len(shard.Blocks)
		}
	}

	// Move a block from the full shard to the empty one
	if len(merkleForest[maxShardIndex].Blocks) > len(merkleForest[minShardIndex].Blocks) {
		blockToMove := merkleForest[maxShardIndex].Blocks[len(merkleForest[maxShardIndex].Blocks)-1]
		merkleForest[maxShardIndex].Blocks = merkleForest[maxShardIndex].Blocks[:len(merkleForest[maxShardIndex].Blocks)-1]
		merkleForest[minShardIndex].Blocks = append(merkleForest[minShardIndex].Blocks, blockToMove)

		// Update Merkle root after moving block
		merkleForest[maxShardIndex].MerkleRoot = updateMerkleRoot(merkleForest[maxShardIndex].Blocks)
		merkleForest[minShardIndex].MerkleRoot = updateMerkleRoot(merkleForest[minShardIndex].Blocks)
	}
}

// Synchronize all shards by updating their Merkle roots
func synchronizeShards() {
	for i := range merkleForest {
		merkleForest[i].MerkleRoot = updateMerkleRoot(merkleForest[i].Blocks)
	}
}

// Synchronize State Across Shards using Merkle Proofs and Authentication
func synchronizeStateAcrossShards(sourceShardIndex, targetShardIndex int) {
	// Get source and target shards
	sourceShard := &merkleForest[sourceShardIndex]
	targetShard := &merkleForest[targetShardIndex]

	// Generate Merkle Proof for the latest block in the source shard
	lastBlockIndex := len(sourceShard.Blocks) - 1
	proof := generateMerkleProof(sourceShardIndex, lastBlockIndex)

	// Cross-shard state transfer: We would transfer necessary data from the source to the target shard.
	// For simplicity, we will transfer the latest block's data and Merkle root to the target shard.

	blockToTransfer := sourceShard.Blocks[lastBlockIndex]

	// Use the Merkle Proof to authenticate the transfer (simplified process)
	isValidProof := validateMerkleProof(sourceShardIndex, lastBlockIndex, proof)
	if !isValidProof {
		fmt.Println("Merkle proof validation failed, aborting state transfer.")
		return
	}

	// Transfer state to target shard (for now, just append the block)
	targetShard.Blocks = append(targetShard.Blocks, blockToTransfer)
	targetShard.MerkleRoot = updateMerkleRoot(targetShard.Blocks)

	// Synchronize the Merkle roots across all shards after state transfer
	synchronizeShards()

	// Optionally, we could add a mechanism to handle consistency between more than two shards.
	// This could involve using authenticated data structures or advanced cryptographic techniques.
}

// Validate Merkle Proof for a given block in a shard
func validateMerkleProof(shardIndex, blockIndex int, proof []string) bool {
	shard := &merkleForest[shardIndex]
	// Recompute the Merkle root by applying the proof in reverse
	calculatedHash := shard.Blocks[blockIndex].Hash
	for _, proofHash := range proof {
		calculatedHash = calculateHashForProof(calculatedHash, proofHash)
	}

	// Compare the recomputed hash with the shard's Merkle root
	return calculatedHash == shard.MerkleRoot
}

// Helper function to calculate hash during Merkle proof validation
func calculateHashForProof(leftHash, rightHash string) string {
	combined := leftHash + rightHash
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}
