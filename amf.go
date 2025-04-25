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

// Adds a block to the shard with fewest blocks (adaptive + dynamic rebalancing + consensus)
func addBlockToShards(data string, validator string) {
	// Smarter shard selection based on load score: fewer blocks + penalty for imbalance
	target := 0
	minScore := len(merkleForest[0].Blocks)
	for i := 1; i < len(merkleForest); i++ {
		blockCount := len(merkleForest[i].Blocks)
		loadScore := blockCount
		if blockCount > maxShardCapacity-1 {
			loadScore += 2 // temporary penalty
		}
		if loadScore < minScore {
			target = i
			minScore = loadScore
		}
	}

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

	if dBFTConsensus(newBlock) {
		shard.Blocks = append(shard.Blocks, newBlock)
		shard.MerkleRoot = updateMerkleRoot(shard.Blocks)

		updateAMQ(target, newBlock.Hash) // ← Add this line

		if len(shard.Blocks) > maxShardCapacity {
			rebalanceShards()
		}

		synchronizeStateAcrossShards(target, (target+1)%len(merkleForest))
	} else {
		fmt.Println("Block rejected by dBFT.")
	}
}

// Merkle Root update for any block list
func updateMerkleRoot(blocks []Block) string {
	if len(blocks) == 0 {
		return ""
	}
	var hashes []string
	for _, block := range blocks {
		hashes = append(hashes, block.Hash)
	}
	for len(hashes) > 1 {
		var newLevel []string
		for i := 0; i < len(hashes); i += 2 {
			right := hashes[i]
			if i+1 < len(hashes) {
				right = hashes[i+1]
			}
			combined := hashes[i] + right
			sum := sha256.Sum256([]byte(combined))
			newLevel = append(newLevel, hex.EncodeToString(sum[:]))
		}
		hashes = newLevel
	}
	return hashes[0]
}

// Merkle Proof generator
func generateMerkleProof(shardIndex, blockIndex int) []string {
	blocks := merkleForest[shardIndex].Blocks
	if blockIndex >= len(blocks) {
		return nil
	}
	var level []string
	for _, block := range blocks {
		level = append(level, block.Hash)
	}
	var proof []string
	index := blockIndex
	for len(level) > 1 {
		var nextLevel []string
		for i := 0; i < len(level); i += 2 {
			left := level[i]
			right := left
			if i+1 < len(level) {
				right = level[i+1]
			}
			combined := left + right
			sum := sha256.Sum256([]byte(combined))
			nextLevel = append(nextLevel, hex.EncodeToString(sum[:]))

			if i == index || i+1 == index {
				sibling := right
				if i+1 == index {
					sibling = left
				}
				proof = append(proof, sibling)
				index = i / 2
			}
		}
		level = nextLevel
	}
	return proof
}

// Rebalance by transferring blocks between shards
func rebalanceShards() {
	var maxShardIndex, minShardIndex int
	maxBlockCount := 0
	minBlockCount := len(merkleForest[0].Blocks)

	for i, shard := range merkleForest {
		count := len(shard.Blocks)
		if count > maxBlockCount {
			maxShardIndex = i
			maxBlockCount = count
		}
		if count < minBlockCount {
			minShardIndex = i
			minBlockCount = count
		}
	}

	if maxShardIndex != minShardIndex && maxBlockCount-minBlockCount > 1 {
		blockToMove := merkleForest[maxShardIndex].Blocks[len(merkleForest[maxShardIndex].Blocks)-1]
		merkleForest[maxShardIndex].Blocks = merkleForest[maxShardIndex].Blocks[:len(merkleForest[maxShardIndex].Blocks)-1]
		merkleForest[minShardIndex].Blocks = append(merkleForest[minShardIndex].Blocks, blockToMove)

		merkleForest[maxShardIndex].MerkleRoot = updateMerkleRoot(merkleForest[maxShardIndex].Blocks)
		merkleForest[minShardIndex].MerkleRoot = updateMerkleRoot(merkleForest[minShardIndex].Blocks)
	}
}

// Updates Merkle roots across all shards
func synchronizeShards() {
	for i := range merkleForest {
		merkleForest[i].MerkleRoot = updateMerkleRoot(merkleForest[i].Blocks)
	}
}

// Cross-shard state sync using Merkle proof
func synchronizeStateAcrossShards(sourceShardIndex, targetShardIndex int) {
	sourceShard := &merkleForest[sourceShardIndex]
	targetShard := &merkleForest[targetShardIndex]

	lastBlockIndex := len(sourceShard.Blocks) - 1
	proof := generateMerkleProof(sourceShardIndex, lastBlockIndex)
	blockToTransfer := sourceShard.Blocks[lastBlockIndex]

	if validateMerkleProof(sourceShardIndex, lastBlockIndex, proof) {
		targetShard.Blocks = append(targetShard.Blocks, blockToTransfer)
		synchronizeShards()
	} else {
		fmt.Println("Merkle proof validation failed, aborting state transfer.")
	}
}

// Merkle Proof validator
func validateMerkleProof(shardIndex, blockIndex int, proof []string) bool {
	leaf := merkleForest[shardIndex].Blocks[blockIndex].Hash
	index := blockIndex
	hash := leaf

	for _, sibling := range proof {
		var combined string
		if index%2 == 0 {
			combined = hash + sibling
		} else {
			combined = sibling + hash
		}
		sum := sha256.Sum256([]byte(combined))
		hash = hex.EncodeToString(sum[:])
		index /= 2
	}

	return hash == merkleForest[shardIndex].MerkleRoot
}

// Not used directly but kept for completeness
func calculateHashForProof(leftHash, rightHash string) string {
	combined := leftHash + rightHash
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// AMQ Filter (simplified): tracks recent block hashes for efficient presence check
type AMQFilter struct {
	HashSet map[string]bool
}

var amqFilters []AMQFilter

// Initialize AMQ filters
func initAMQFilters() {
	for i := 0; i < shardCount; i++ {
		amqFilters = append(amqFilters, AMQFilter{HashSet: make(map[string]bool)})
	}
}

// Update AMQ when block added
func updateAMQ(shardIndex int, hash string) {
	amqFilters[shardIndex].HashSet[hash] = true
}

// Check block presence using AMQ
func isInAMQ(shardIndex int, hash string) bool {
	return amqFilters[shardIndex].HashSet[hash]
}

// Probabilistic Merkle proof compression (truncate each hash to first 8 chars)
func compressMerkleProof(proof []string) []string {
	var compressed []string
	for _, h := range proof {
		if len(h) >= 8 {
			compressed = append(compressed, h[:8])
		}
	}
	return compressed
}

// Cryptographic accumulator snapshot (accumulated XOR of hashes)
func getAccumulatorSnapshot(shardIndex int) string {
	acc := make([]byte, 32)
	for _, block := range merkleForest[shardIndex].Blocks {
		hashBytes, _ := hex.DecodeString(block.Hash)
		for i := range acc {
			acc[i] ^= hashBytes[i]
		}
	}
	return hex.EncodeToString(acc)
}
