package cmd

import "fmt"

// TODO: cleanup mess with modes

func Execute(cfg *Config, bleed *Bleed) error {
	fmt.Println("Executing")
	fmt.Println("Config")
	fmt.Printf("Audio %v\n", cfg.Audio)
	fmt.Printf("Output %v\n", cfg.Output)
	fmt.Printf("Commands %v\n", cfg.Commands)
	fmt.Printf("Symbols %v\n", cfg.Symbols)

	fmt.Println("Bleed")
	fmt.Printf("Include %v\n", bleed.Include)
	fmt.Printf("Sequence %v\n", bleed.Sequence)

	for k, v := range bleed.Sequence {
		fmt.Printf("Sequence [%s]\n", k)
		fmt.Printf("Sequence args %v\n", v.Args)
		fmt.Printf("Sequence reps %v\n", v.Repeat)
		fmt.Printf("Sequence content %v\n", v.Content)
	}

	bleeder := NewBleeder(cfg, bleed)
	fmt.Printf("Bleeder %v", bleeder)

	return nil
}

type Mode interface {
	Run(args ...any) error
}

type ModePlay struct{}

func (m ModePlay) Run(args ...any) error {
	file := args[0].(string) // unsafe cast
	fmt.Printf("Playing %v\n", file)
	return nil
}

type ModeSend struct{}

func (m ModeSend) Run(args ...any) error {
	fmt.Printf("Sending %v\n", args...)
	return nil
}
