package main

import (
	"os"

	"github.com/Gulliby-development/youngkit/internal/base"
	"github.com/Gulliby-development/youngkit/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], base.StdIO{}))
}
