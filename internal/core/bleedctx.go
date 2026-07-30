package core

import (
	"bleeder/internal/ir"
	"fmt"
	"io"
	"time"
)

type BleedContext struct {
	bleed    *Bleed
	bleeder  *Bleeder
	commands chan BleedContextRequest
}

type BleedContextRequest struct {
	Type string
	Seq  string
	Vars string
	Resp chan BleedContextResponse
}

type BleedContextResponse struct {
	Error error
	Info  string
}

func NewBleedContext(bleed *Bleed) *BleedContext {
	return &BleedContext{
		bleed:    bleed,
		bleeder:  NewBleeder(bleed),
		commands: make(chan BleedContextRequest, 10),
	}
}

func (ctx *BleedContext) Play(seq, vars string) error {
	resp := make(chan BleedContextResponse, 1)
	ctx.commands <- BleedContextRequest{
		Type: "PLAY",
		Seq:  seq,
		Vars: vars,
		Resp: resp,
	}
	r := <-resp
	return r.Error
}

func (ctx *BleedContext) Stop() error {
	resp := make(chan BleedContextResponse, 1)
	ctx.commands <- BleedContextRequest{
		Type: "STOP",
		Resp: resp,
	}
	r := <-resp
	return r.Error
}

func (ctx *BleedContext) Sync() error {
	resp := make(chan BleedContextResponse, 1)
	ctx.commands <- BleedContextRequest{
		Type: "SYNC",
		Resp: resp,
	}
	r := <-resp
	return r.Error
}

func (ctx *BleedContext) Info() string {
	resp := make(chan BleedContextResponse, 1)
	ctx.commands <- BleedContextRequest{
		Type: "INFO",
		Resp: resp,
	}
	r := <-resp
	return r.Info
}

func (ctx *BleedContext) Run(w io.Writer) {
	var (
		irp       *ir.Program
		playing   bool
		seq       string
		vars      string
		startTime time.Time
		sentIdx   int // Track which instructions we've already sent
	)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case cmd := <-ctx.commands:
			switch cmd.Type {
			case "PLAY":
				newIrp, err := ctx.bleeder.GenSeqIR(cmd.Seq, cmd.Vars)
				if err != nil {
					cmd.Resp <- BleedContextResponse{Error: err}
					continue
				}
				irp = newIrp
				seq, vars, playing = cmd.Seq, cmd.Vars, true
				startTime = time.Now()
				sentIdx = 0
				cmd.Resp <- BleedContextResponse{Error: nil}

			case "STOP":
				if !playing {
					cmd.Resp <- BleedContextResponse{Error: fmt.Errorf("is not playing")}
					continue
				}
				playing = false
				cmd.Resp <- BleedContextResponse{Error: nil}

			case "SYNC":
				newBleed, err := LoadBleed(ctx.bleed.Meta.Path)
				if err != nil {
					cmd.Resp <- BleedContextResponse{Error: err}
					continue
				}
				ctx.bleed = newBleed
				ctx.bleeder = NewBleeder(newBleed)
				if playing && seq != "" {
					// Save current position
					currentElapsed := time.Since(startTime).Seconds()

					newIrp, err := ctx.bleeder.GenSeqIR(seq, vars)
					if err != nil {
						cmd.Resp <- BleedContextResponse{Error: err}
						continue
					}
					irp = newIrp

					// Find where we should be in the new IR
					instructions := irp.Instructions()
					newSentIdx := 0
					for i, ins := range instructions {
						if ins.Time > currentElapsed {
							break
						}
						newSentIdx = i + 1
					}
					sentIdx = newSentIdx

					// Adjust startTime to maintain current position
					startTime = time.Now().Add(-time.Duration(currentElapsed * float64(time.Second)))
				}
				cmd.Resp <- BleedContextResponse{Error: nil}

			case "INFO":
				info := fmt.Sprintf("seq=%s playing=%v",
					seq, playing)
				cmd.Resp <- BleedContextResponse{Info: info, Error: nil}
			}

		case <-ticker.C:
			if !playing || irp == nil {
				continue
			}

			elapsed := time.Since(startTime).Seconds()
			instructions := irp.Instructions()

			// Send any instructions that should start now
			for sentIdx < len(instructions) {
				ins := instructions[sentIdx]
				if ins.Time > elapsed {
					break // Future instruction, wait
				}
				// Send this instruction
				fmt.Fprintln(w, ins.Serialize())
				sentIdx++
			}

			// Check if we need to loop
			if sentIdx >= len(instructions) && elapsed >= irp.Duration() {
				sentIdx = 0
				startTime = time.Now()
			}
		}
	}
}

func (ctx *BleedContext) Generate(name, vars string, w io.Writer) error {
	irp, err := ctx.bleeder.GenSeqIR(name, vars)
	if err != nil {
		return err
	}

	for _, ins := range irp.Instructions() {
		fmt.Fprintln(w, ins.Serialize())
	}

	return nil
}
