package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Hashing
func calculateHash(block Block) string {
	record := fmt.Sprintf("%d%s%s%s%d%s", block.Index, block.Timestamp, block.Data, block.PrevHash, block.Nonce, block.Validator)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

func isValidHash(hash string, difficulty int) bool {
	prefix := ""
	for i := 0; i < difficulty; i++ {
		prefix += "0"
	}
	return hash[:difficulty] == prefix
}
