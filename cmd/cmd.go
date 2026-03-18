package cmd

import "fmt"

func Execute(cfg *Config, bleed *Bleed) error {
	fmt.Println("Executing")
	fmt.Println("Config")
	fmt.Printf("Audio %v\n", cfg.Audio)
	fmt.Printf("Output %v\n", cfg.Output)
	fmt.Printf("Mapping %v\n", cfg.Mapping)

	fmt.Println("Bleed")
	fmt.Printf("Include %v\n", bleed.Include)
	fmt.Printf("Sequence %v\n", bleed.Sequence)

	return nil
}
