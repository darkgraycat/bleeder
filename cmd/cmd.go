package cmd

import "fmt"

type CmdArgs struct {
	BleedPath string
	CfgPath   string
	Seq       string
	Raw       string
}

type Cmd func(args *CmdArgs) error

// Command to play specified .bleed.toml file
func CmdPlay(args *CmdArgs) error {
	fmt.Printf("[PLAY] %v\n", args)
	cfg, err := LoadConfig(args.CfgPath)
	if err != nil {
		return fmt.Errorf("Config - %v", err)
	}
	bleed, err := LoadBleed(args.BleedPath)
	if err != nil {
		return fmt.Errorf("Bleed - %v", err)
	}
	bleeder, err := NewBleeder(cfg).Bleed(bleed)
	if err != nil {
		return err
	}

	fmt.Printf("[PLAY] GetMainIR()\n")
	ir, err := bleeder.GetMainIR()
	if err != nil {
		return err
	}
	// TODO: render IR using player
	fmt.Printf("IR %v\n", ir)

	return nil
}

// Start application in daemon mode listening
func CmdListen(args *CmdArgs) error {
	fmt.Printf("[LISTEN] %v\n", args)
	return nil
}

// Send partial sequence data to play
func CmdSend(args *CmdArgs) error {
	fmt.Printf("[SEND] %v\n", args)
	return nil
}

