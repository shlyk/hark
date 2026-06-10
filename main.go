package main

import (
	"os"

	"github.com/shlyk/hark/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
