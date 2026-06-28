package experiments

import (
	"bleeder/cmd"
	"bleeder/internal/bleeder"
	"bleeder/internal/player"
	"bleeder/internal/shared/logs"
	"fmt"
)

func Run() {
	logs.SetLogLevel(logs.DEBUG) // debug
	err := runExp1()
	if err != nil {
		logs.Error("Error: %v\n", err)
	}
}

func runExp1() error {
	fmt.Printf("Experiment 1\n")
	cfg, err := cmd.LoadConfig("./config.toml")
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	bleed, err := bleeder.LoadBleed("./experiments/test.toml")
	if err != nil {
		return fmt.Errorf("loading bleed: %w", err)
	}

	b := bleeder.NewBleeder(bleed)

	irp, err := b.GenMainIR()
	if err != nil {
		return fmt.Errorf("getting IR: %w", err)
	}

	fmt.Printf("IR - %v\n", irp)

	p := player.NewWAVPlayer(cfg.Audio.SampleRate, cfg.Audio.Channels)
	err = p.Play(irp, 0, irp.Length())
	if err != nil {
		return err
	}
	return nil
}
