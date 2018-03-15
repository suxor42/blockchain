package main

import (
	"./blockchain"
	"fmt"
)

func main() {
	bc := blockchain.NewBlockChain()
	bc.AddBlock("Block 1")
	bc.AddBlock("Block 2")

	for _, block := range blockchain.Blocks() {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()
	}
	return
}
