package cmd

import (
	"bleeder/internal/utils"
	"fmt"
)

type Cmd func(args *utils.Args) error

func CmdPlay(args *utils.Args) error {
	cfg, err := LoadConfig(args.At(2))
	bleed, err := LoadBleed(args.At(2))
	if err != nil {
		return err
	}
	return Test(cfg, bleed)
}

func CmdSend(args *utils.Args) error {
	fmt.Printf("[SEND] %v\n", args)
	return nil
}

func CmdServe(args *utils.Args) error {
	fmt.Printf("[SERVE] %v\n", args)
	return nil
}

func Test(cfg *Config, bleed *Bleed) error {
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
