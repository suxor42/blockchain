package cli

import "flag"

type CLI struct {
	commands map[string] command
	address string
}

type command struct {
	function func()
	flagSet *flag.FlagSet
}