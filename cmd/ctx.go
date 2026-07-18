package cmd

import (
	"bleeder/internal/core"
	"bleeder/internal/ir"
	"bleeder/internal/player"
	"fmt"
	"sync"
)

type CmdContext struct {
	cfg       *Config
	bleed     *core.Bleed
	bleeder   *core.Bleeder
	irp       *ir.Program
	isPlaying bool
	mu        sync.Mutex
}

func NewCmdContext(cfg *Config, bleed *core.Bleed) *CmdContext {
	return &CmdContext{
		cfg:     cfg,
		bleed:   bleed,
		bleeder: core.NewBleeder(bleed),
	}
}

func (ctx *CmdContext) Play(name, vars string) error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	irp, err := ctx.bleeder.GenSeqIR(name, vars)
	if err != nil {
		return fmt.Errorf("generating %q %q: %w", name, vars, err)
	}
	// TODO: do proper design of renderers
	p := player.NewWAVPlayer(ctx.cfg.Audio.SampleRate, ctx.cfg.Audio.Channels)
	err = p.Play(irp, 0, irp.Length())
	if err != nil {
		return fmt.Errorf("playing %q %q: %w", name, vars, err)
	}
	ctx.isPlaying = true
	ctx.irp = irp
	return nil
}

func (ctx *CmdContext) Stop() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return nil
}

func (ctx *CmdContext) Update() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return nil
}

func (ctx *CmdContext) Info() string {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return ""
}
