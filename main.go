package main

import (
	"bleeder/cmd"
	"fmt"
	"os"
)

func main() {
	args := os.Args
	fmt.Println("CLI arguments are", args)

	cfg, err := cmd.LoadConfig("config.toml")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading config file", err)
		os.Exit(1)
	}

	bleed, err := cmd.LoadBleed("experiments/song_a.toml")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading bleed file", err)
		os.Exit(1)
	}

	err = cmd.Execute(cfg, bleed)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Runtime error", err)
		os.Exit(1)
	}
}
