package cmd

import "bleeder/internal/ir"

type Bleeder struct {
	cfg     *Config
	bleed   *Bleed
	program *ir.Program[any]
}

func NewBleeder(cfg *Config, bleed *Bleed) *Bleeder {
	// TODO: parse bleed
	return &Bleeder{
		cfg:   cfg,
		bleed: bleed,
	}
}
