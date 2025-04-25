package main

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"time"
)

// Constants for CAP behavior
const (
	Consistency = iota
	Availability
	PartitionTolerance
)

// Constants for Byzantine Fault Tolerance
const (
	TrustThreshold = 0.6 // Minimum trust level to consider validator's vote
)

var currentState = Consistency // Can be updated dynamically
var retryCount = 0

// Validators pool

// CAPOrchestrator orchestrates CAP tradeoffs.
func CAPOrchestrator() {
	predictNetworkPartition()
	switch currentState {
	case Consistency:
		fmt.Println("System is in Consistency mode.")
		ensureConsistency()
	case Availability:
		fmt.Println("System is in Availability mode.")
		ensureAvailability()
	case PartitionTolerance:
		fmt.Println("System is in Partition Tolerance mode.")
		ensurePartitionTolerance()
	default:
		fmt.Println("Unknown mode. Defaulting to Consistency.")
		ensureConsistency()
	}
}

// --- Core Modes ---

func ensureConsistency() {
	fmt.Println("Ensuring strong consistency...")
	synchronizeShards()
	applyVectorClocks()
}

func ensureAvailability() {
	fmt.Println("Allowing writes during partition...")
	markPendingUpdates()
}

func ensurePartitionTolerance() {
	fmt.Println("Handling partitions with retry and timeout...")
	retrySynchronization()
}

func markPendingUpdates() {
	fmt.Println("Tagging updates as pending for later sync.")
}

func retrySynchronization() {
	retryCount++
	timeout := adaptiveTimeout()
	fmt.Printf("Retry #%d with timeout %v\n", retryCount, timeout)
	time.Sleep(timeout)
}

// --- Adaptive and Advanced Features ---

func adaptiveTimeout() time.Duration {
	latency := measureNetworkLatency()
	if latency > 200 {
		return 5 * time.Second
	}
	return 2 * time.Second
}

func measureNetworkLatency() int {
	// Simulated latency in milliseconds
	return rand.Intn(300)
}

func predictNetworkPartition() {
	if rand.Float64() < 0.3 {
		currentState = PartitionTolerance
		fmt.Println("Predicted network partition: switching mode.")
	} else if rand.Float64() < 0.5 {
		currentState = Availability
		fmt.Println("Network unstable: favoring availability.")
	} else {
		currentState = Consistency
		fmt.Println("Network stable: favoring consistency.")
	}
}

// --- Vector Clock Simulation ---
var vectorClock = map[string]int{
	"Node1": 0, // Vector clock for Node1
	"Node2": 0, // Vector clock for Node2
	"Node3": 0, // Vector clock for Node3
}

// applyVectorClocks simulates vector clocks for causal consistency.
func applyVectorClocks() {
	fmt.Println("Applying vector clocks for causal consistency.")

	// Simulate an update from Node1
	vectorClock["Node1"]++
	fmt.Printf("Node1's vector clock: %v\n", vectorClock["Node1"])

	// Simulate communication between Node1 and Node2
	synchronizeClocks("Node1", "Node2")
	fmt.Printf("After sync, Node1's vector clock: %v, Node2's vector clock: %v\n", vectorClock["Node1"], vectorClock["Node2"])

	// Simulate an update from Node3
	vectorClock["Node3"]++
	fmt.Printf("Node3's vector clock: %v\n", vectorClock["Node3"])

	// Simulate communication between Node2 and Node3
	synchronizeClocks("Node2", "Node3")
	fmt.Printf("After sync, Node2's vector clock: %v, Node3's vector clock: %v\n", vectorClock["Node2"], vectorClock["Node3"])
}

// synchronizeClocks simulates synchronization between two nodes' vector clocks.
func synchronizeClocks(node1, node2 string) {
	// Take the element-wise max of the two clocks to simulate synchronization
	if vectorClock[node1] > vectorClock[node2] {
		vectorClock[node2] = vectorClock[node1]
	} else if vectorClock[node2] > vectorClock[node1] {
		vectorClock[node1] = vectorClock[node2]
	}
}

func detectConflicts() bool {
	return rand.Float64() < 0.2 // 20% simulated conflict rate
}

func resolveConflicts() {
	if detectConflicts() {
		fmt.Println("Conflict detected! Applying entropy-based resolution...")
		probabilisticResolution()
	} else {
		fmt.Println("No conflict detected.")
	}
}

func probabilisticResolution() {
	prob := rand.Float64()
	if prob < 0.5 {
		fmt.Println("Resolution: Accept higher entropy state.")
	} else {
		fmt.Println("Resolution: Merge divergent states.")
	}
}

// --- Byzantine Fault Tolerance (BFT) ---

func validateBFT(block Block) bool {
	var totalTrust, approvedTrust float64
	var totalVotes, maliciousVotes int

	var validators = map[string]*ValidatorProfile{
		"Validator1": {Trust: 0.9, History: 3, StakeLevel: 3, LastPing: time.Now(), PublicKey: "pk1"},
		"Validator2": {Trust: 0.7, History: 2, StakeLevel: 2, LastPing: time.Now(), PublicKey: "pk2"},
		"Validator3": {Trust: 0.4, History: 1, StakeLevel: 1, LastPing: time.Now().Add(-2 * time.Minute), PublicKey: "pk3"},
		"Validator4": {Trust: 0.2, History: 0, StakeLevel: 0, LastPing: time.Now(), PublicKey: "pk4"},
	}

	for id, v := range validators {
		if v.Trust < TrustThreshold || v.StakeLevel < 1 {
			continue
		}
		if time.Since(v.LastPing) > time.Minute*2 { // Example: max ping timeout for auth
			continue
		}
		if !verifyZKProof(v.PublicKey) {
			continue
		}

		randomInput := fmt.Sprintf("%s:%s", id, block.Hash)
		randomHash := sha256.Sum256([]byte(randomInput))
		randomScore := float64(randomHash[0]) / 255.0

		trustFactor := v.Trust * 0.7
		historyBoost := float64(v.History) * 0.05
		randomBoost := randomScore * 0.25

		effectiveScore := trustFactor + historyBoost + randomBoost
		vote := effectiveScore > 0.6

		totalTrust += v.Trust
		totalVotes++

		if vote {
			approvedTrust += v.Trust
			v.History++
		} else {
			maliciousVotes++
			v.History--
			if v.History < -3 {
				v.Trust *= 0.9 // Penalize malicious behavior
			}
		}
	}

	if totalVotes == 0 {
		return false
	}

	avgTrust := totalTrust / float64(len(validators))
	dynamicThreshold := 0.5 + (1-avgTrust)*0.2
	ratio := approvedTrust / totalTrust

	if float64(maliciousVotes)/float64(totalVotes) > 0.6 {
		fmt.Println("Byzantine Fault: Majority of votes are malicious, rejecting consensus.")
		return false
	}

	return ratio >= dynamicThreshold
}
