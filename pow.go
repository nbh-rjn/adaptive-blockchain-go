package main

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
