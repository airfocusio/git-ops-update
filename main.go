package main

import (
	"os"

	"github.com/choffmeister/git-ops-update/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
