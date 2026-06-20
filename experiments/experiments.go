package experiments

import (
	"bleeder/internal/bleeder"
	"bleeder/internal/shared/logs"
	"fmt"
)

func Run() {
	logs.SetLogLevel(2) // debug
	err := runExp1()
	if err != nil {
		fmt.Printf("ERROR: %v", err)
	}
}

func runExp1() error {
	fmt.Printf("Experiment 1\n")
	bleed, err := bleeder.LoadBleed("./experiments/test.toml")
	if err != nil {
		return fmt.Errorf("bleed - %v", err)
	}

	b := bleeder.NewBleeder(bleed)

	irp, err := b.GenMainIR()
	if err != nil {
		return fmt.Errorf("IR - %v", err)
	}

	fmt.Printf("IR - %v", irp)

	return nil
}
