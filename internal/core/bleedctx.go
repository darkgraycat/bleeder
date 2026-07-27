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
		pos       float64
		seq       string
		vars      string
		startTime time.Time
		times     []float64
		timeIdx   int
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
				seq, vars, pos, playing = cmd.Seq, cmd.Vars, 0.0, true
				times, timeIdx, startTime = irp.Times(), 0, time.Now()
				cmd.Resp <- BleedContextResponse{Error: nil}
			case "STOP":
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
					times = irp.Times()
				}
				cmd.Resp <- BleedContextResponse{Error: nil}
			case "INFO":
				info := fmt.Sprintf("seq=%s pos=%.2f playing=%v",
					seq, pos, playing)
				cmd.Resp <- BleedContextResponse{Info: info, Error: nil}
			}
		case <-ticker.C:
			if !playing || irp == nil || timeIdx >= len(times) {
				if playing && irp != nil && timeIdx >= len(times) {
					elapsed := time.Since(startTime).Seconds()
					duration := irp.Duration()
					if duration > elapsed {
						sleepDur := time.Duration((duration - elapsed) * float64(time.Second))
						time.Sleep(sleepDur)
					}
					timeIdx = 0
					pos = 0.0
					startTime = time.Now()
				}
				continue
			}
			t := times[timeIdx]
			elapsed := time.Since(startTime).Seconds()
			if t > elapsed {
				continue
			}
			chunk := irp.AtTime(t)
			for _, ins := range chunk {
				fmt.Fprintf(w, "[+%.3fs] %v\n", time.Since(startTime).Seconds(), ins)
			}
			pos = t
			timeIdx++
		}
	}
}
