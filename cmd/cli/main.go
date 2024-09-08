package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a subcommand")
		os.Exit(1)
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "version":
		printVersion()
	default:
		fmt.Println("Invalid subcommand")
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Println("0.01")
}
