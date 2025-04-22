package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Validator trust scores (simulated)
var validatorScores = map[string]float64{
	"Validator1": 0.9,
	"Validator2": 0.7,
	"Validator3": 0.4,
	"Validator4": 0.2,
}

const minTrustScore = 0.5     // Minimum trust needed to approve a block
const approvalThreshold = 0.6 // 60% of total trust votes

func dBFTConsensus(block Block) bool {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("Applying hybrid consensus on block. Block Data:", block.Data)

	var totalTrust float64
	var approvedTrust float64

	for validator, score := range validatorScores {
		// Skip Byzantine/unreliable nodes
		if score < minTrustScore {
			fmt.Printf("%s skipped (low trust: %.2f)\n", validator, score)
			continue
		}

		// Simulate unresponsiveness (10% chance)
		if rand.Float64() < 0.1 {
			fmt.Printf("%s did not respond\n", validator)
			continue
		}

		vote := rand.Float64() <= score // higher trust = higher chance to approve
		totalTrust += score

		if vote {
			fmt.Printf("%s voted ✅ (score: %.2f)\n", validator, score)
			approvedTrust += score
		} else {
			fmt.Printf("%s voted ❌ (score: %.2f)\n", validator, score)
		}
	}

	if totalTrust == 0 {
		fmt.Println("No valid validator responses.")
		return false
	}

	approvalRatio := approvedTrust / totalTrust
	fmt.Printf("Approval ratio: %.2f (threshold: %.2f)\n", approvalRatio, approvalThreshold)

	return approvalRatio >= approvalThreshold
}
