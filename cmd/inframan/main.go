package main

import (
	"os"

	"github.com/iivel-inc/inframan/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}

