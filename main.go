package main

import (
	"./cli"
)

func main() {
	c := cli.CreateCli()
	c.Run()
	return
}
