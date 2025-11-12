package main

import (
	"fmt"
	"os"

	"github.com/v-egorov/service-boilerplate/migration-orchestrator/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
