package main

import (
	"os"

	"github.com/fantods/gjq/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
