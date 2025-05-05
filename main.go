package main

import (
	"fmt"
	"os"

	cli "github.com/Solexma/bloatjack/internal/cli"
)

var Version string

func main() {
	if Version != "" {
		cli.Version = Version
	}

	if err := cli.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
