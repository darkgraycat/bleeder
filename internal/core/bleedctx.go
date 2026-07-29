package core

import (
	"bleeder/internal/ir"
	"bleeder/internal/renderer"
	"fmt"
	"io"
	"log"
	"os"
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
		pos       float64
		seq       string
		vars      string
		startTime time.Time
	)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	wr := renderer.NewWAVRenderer(44010, 1)
	wr.Start(w)

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
				seq, vars, pos, playing = cmd.Seq, cmd.Vars, 0.0, true
				startTime = time.Now()
				cmd.Resp <- BleedContextResponse{Error: nil}
			case "STOP":
				if !playing {
					cmd.Resp <- BleedContextResponse{Error: fmt.Errorf("is not playing")}
					continue
				}
				log.Println("[DEBUG] STOP: setting playing=false")
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
					newIrp, err := ctx.bleeder.GenSeqIR(seq, vars)
					if err != nil {
						cmd.Resp <- BleedContextResponse{Error: err}
						continue
					}
					irp = newIrp
				}
				cmd.Resp <- BleedContextResponse{Error: nil}
			case "INFO":
				info := fmt.Sprintf("seq=%s pos=%.2f playing=%v",
					seq, pos, playing)
				cmd.Resp <- BleedContextResponse{Info: info, Error: nil}
			}
		case <-ticker.C:
			// Always write samples - silence if not playing
			chunkDuration := 0.01 // 10ms

			var activeInstructions []*ir.Instruction

			if playing && irp != nil {
				elapsed := time.Since(startTime).Seconds()
				duration := irp.Duration()

				// Check if we need to loop
				if elapsed >= duration {
					log.Println("[DEBUG] LOOP: restarting sequence")
					pos = 0.0
					startTime = time.Now()
					elapsed = 0.0
				}

				// Get all instructions that are currently active
				// (started and not finished yet)
				for _, ins := range irp.Instructions() {
					insStart := ins.Time
					insEnd := ins.Time + ins.Dur
					if elapsed >= insStart && elapsed < insEnd {
						activeInstructions = append(activeInstructions, ins)
					}
				}

				if len(activeInstructions) > 0 {
					log.Printf("[DEBUG] Playing %d notes at Time=%.3f\n", len(activeInstructions), elapsed)
				}
				pos = elapsed
			}

			// Write chunk (silence if activeInstructions is empty)
			wr.WriteChunk(chunkDuration, pos, activeInstructions, w)
			if f, ok := w.(*os.File); ok {
				f.Sync()
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
