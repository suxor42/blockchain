package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"github.com/suxor42/blockchain/blockchain"
)

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  addblock -data BLOCK_DATA - add a block to the blockchain")
	fmt.Println("  printchain - print all the blocks of the blockchain")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func CreateCli() *CLI{
	return &CLI{
		make(map[string]command),
		"",
		}
}

func (cli *CLI) Run() {
	cli.validateArgs()

	// Adding sub commands
	cli.commands["printchain"] = command{cli.printChain, flag.NewFlagSet("printchain", flag.ExitOnError)}
	cli.commands["createchain"] = command{cli.printChain, flag.NewFlagSet("createchain", flag.ExitOnError)}
	cli.commands["addblock"] = command{cli.addBlock, flag.NewFlagSet("addblock", flag.ExitOnError)}
	cli.commands["balance"] = command{cli.balance, flag.NewFlagSet("balance", flag.ExitOnError)}

	// Adding params for printchain
	printChainAddress := cli.commands["printchain"].flagSet.String("address", "", "Address of the chain")

	// Adding params for createchain
	createChainAddress := cli.commands["createchain"].flagSet.String("address", "", "The address to send genesis block reward to")

	// Adding params for balance
	balanceAddress := cli.commands["balance"].flagSet.String("address", "", "The address to get balance for")


	switch os.Args[1] {
	case "createchain":
		err := cli.commands[os.Args[1]].flagSet.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
		cli.address = *createChainAddress
	case "printchain":
		err := cli.commands[os.Args[1]].flagSet.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
		cli.address = *printChainAddress
	case "balance":
		err := cli.commands[os.Args[1]].flagSet.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
		cli.address = *balanceAddress
	default:
		cli.printUsage()
		os.Exit(1)
	}

	for _, command := range cli.commands {
		if command.flagSet.Parsed() {
			command.function()
		}
	}

}

func (cli *CLI) addBlock() {
	bc := blockchain.NewBlockChain(cli.address)
	defer bc.Db.Close()
	fmt.Println("Success!")
}

func (cli *CLI) printChain() {
	bc := blockchain.NewBlockChain(cli.address)
	defer bc.Db.Close()
	iterator := bc.Iterator()
	for {
		block := iterator.Next()
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Transactions)
		fmt.Printf("Hash: %x\n", block.Hash)
		proofOfWork := blockchain.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(proofOfWork.Validate()))
		fmt.Printf("Nonce: %d\n", block.Nonce)
		fmt.Println()
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) createBlockchain() {
	bc := blockchain.NewBlockChain(cli.address)
	defer bc.Db.Close()
	fmt.Println("Done")
}

func (cli *CLI) balance() {
	bc := blockchain.NewBlockChain(cli.address)
	defer bc.Db.Close()
	balance := bc.Balance(cli.address)
	fmt.Printf("Balance for %s: %d\n", cli.address, balance)
}