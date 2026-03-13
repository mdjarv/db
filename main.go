// Package main is the entry point for db.
package main

import (
	"os"

	"github.com/mdjarv/db/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
