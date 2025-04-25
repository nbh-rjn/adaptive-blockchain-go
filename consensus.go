package main

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"time"
)

// Extended validator profile
type ValidatorProfile struct {
	Trust      float64
	History    int
	Location   string
	PublicKey  string
	StakeLevel int
	LastPing   time.Time
}

var validators = map[string]*ValidatorProfile{
	"Validator1": {Trust: 0.9, History: 3, Location: "US", PublicKey: "pk1", StakeLevel: 3, LastPing: time.Now()},
	"Validator2": {Trust: 0.7, History: 2, Location: "EU", PublicKey: "pk2", StakeLevel: 2, LastPing: time.Now()},
	"Validator3": {Trust: 0.4, History: 1, Location: "AS", PublicKey: "pk3", StakeLevel: 1, LastPing: time.Now().Add(-2 * time.Minute)},
	"Validator4": {Trust: 0.2, History: 0, Location: "AF", PublicKey: "pk4", StakeLevel: 0, LastPing: time.Now()},
}

const baseThreshold = 0.5
const authTimeout = 90 * time.Second

// External proof interface
type ExternalProofProvider interface {
	VerifyZK(publicKey string) bool
	RunMPC(nodeCount int) bool
}

type SimulatedProofProvider struct{}

func (p *SimulatedProofProvider) VerifyZK(publicKey string) bool {
	return verifyZKProof(publicKey)
}

func (p *SimulatedProofProvider) RunMPC(nodeCount int) bool {
	return simulateMPC(nodeCount)
}

var proofProvider ExternalProofProvider = &SimulatedProofProvider{}

func mineBlock(block Block) int {
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

func dBFTConsensus(block Block) bool {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("Hybrid Consensus: dBFT + PoW randomness")

	var totalTrust, approvedTrust float64
	var trustValues []float64
	var maliciousVotes int
	var totalVotes int

	for id, v := range validators {
		if v.Trust < 0.3 || v.StakeLevel < 1 {
			fmt.Printf("%s skipped (low trust/stake)\n", id)
			continue
		}
		if time.Since(v.LastPing) > authTimeout {
			fmt.Printf("%s failed auth (stale ping)\n", id)
			continue
		}
		if !proofProvider.VerifyZK(v.PublicKey) {
			fmt.Printf("%s failed cryptographic check\n", id)
			continue
		}

		randomInput := fmt.Sprintf("%s:%s", id, block.Hash)
		randomHash := sha256.Sum256([]byte(randomInput))
		randomScore := float64(randomHash[0]) / 255.0
		vrfOutput := fmt.Sprintf("%x", randomHash)

		trustFactor := v.Trust * 0.7
		historyBoost := float64(v.History) * 0.05
		randomBoost := randomScore * 0.25

		effectiveScore := trustFactor + historyBoost + randomBoost
		vote := effectiveScore > 0.6

		stakeWeight := float64(v.StakeLevel) / 3.0
		weightedTrust := v.Trust * stakeWeight

		totalTrust += v.Trust
		trustValues = append(trustValues, v.Trust)
		totalVotes++

		if vote {
			fmt.Printf("%s voted ✅ (score: %.2f, vrf: %s)\n", id, effectiveScore, vrfOutput[:8])
			approvedTrust += weightedTrust
			v.History++
		} else {
			fmt.Printf("%s voted ❌ (score: %.2f, vrf: %s) ❌ REJECTED\n", id, effectiveScore, vrfOutput[:8])
			maliciousVotes++
			v.History--
			if v.History < -3 {
				v.Trust *= 0.9
			}
		}
	}

	if totalTrust == 0 {
		fmt.Println("No validators responded.")
		return false
	}

	avgTrust := average(trustValues)
	dynamicThreshold := baseThreshold + (1-avgTrust)*0.2
	ratio := approvedTrust / totalTrust

	fmt.Printf("Approval Ratio: %.2f | Required: %.2f\n", ratio, dynamicThreshold)

	if totalVotes > 0 && float64(maliciousVotes)/float64(totalVotes) > 0.6 {
		fmt.Println("Consensus failed: majority of validators likely malicious.")
		return false
	}

	if proofProvider.RunMPC(totalVotes) {
		fmt.Println("MPC agreement confirmed.")
	} else {
		fmt.Println("MPC failure.")
		return false
	}

	return ratio >= dynamicThreshold
}

// Simulated MPC agreement
func simulateMPC(validators int) bool {
	return rand.Float64() < 0.95
}

// Simulated ZK proof verification
func verifyZKProof(publicKey string) bool {
	hash := sha256.Sum256([]byte(publicKey))
	value := int(hash[0]) + int(hash[1]) + int(hash[2])
	return value%10 < 9
}

func average(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}
