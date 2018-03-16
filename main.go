package main

import (
	"github.com/suxor42/blockchain/cli"
)

func main() {
	c := cli.CreateCli()
	c.Run()
	return
}
