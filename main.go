package main

import (
	"os"

	"github.com/waynenilsen/crumbler/cmd/crumbler"
)

func main() {
	if err := crumbler.Execute(); err != nil {
		os.Exit(1)
	}
}
