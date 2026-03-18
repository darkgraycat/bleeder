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

	for k, v := range bleed.Sequence {
		fmt.Printf("Sequence [%s]\n", k)
		fmt.Printf("Sequence args %v\n", v.Args)
		fmt.Printf("Sequence reps %v\n", v.Repeat)
		fmt.Printf("Sequence content %v\n", v.Content)
	}

	return nil
}
