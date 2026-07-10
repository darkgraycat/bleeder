package cmd

import (
	"bleeder/internal/core"
	"bleeder/internal/ir"
	"sync"
)

type CmdContext struct {
	cfg       *Config
	path      string
	bleeder   *core.Bleeder
	irp       *ir.Program
	isPlaying bool
	mu        sync.Mutex
}

func (ctx *CmdContext) cmdPlay(seqName, vars string) error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return nil
}

func (ctx *CmdContext) cmdStop() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return nil
}

func (ctx *CmdContext) cmdUpdate() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return nil
}

func (ctx *CmdContext) cmdInfo() string {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return ""
}
