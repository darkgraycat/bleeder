// TODO:
// Add property Mark int (public)
// Add method SetMark() int
// Add method GetMark() *Instruction
//
// BUT: I should then INSERT elements
// bcause with current implementation its just for read
// SO:
// I can try implementing Cursor instead
// and keep track of current program state
//
// BTW:
// Do we really need mark functionality?

package experiments

import (
	"bleeder/internal/bleeder"
	"bleeder/internal/ir"
	"bleeder/internal/player"
	"bleeder/internal/shared/logs"
)

type TestSeq struct {
	Content string            `toml:"content"`
	Vars    map[string]string `toml:???`
}

func Run() {
	logs.SetLogLevel(2) // debug

}

func runExp1() {
	context := &bleeder.ParserContext{
		ResolveFunc: func(name string, args []string) (*ir.Program, error) {
			return bleeder.ParseContent("> e2 2 |+7 |+7", nil)
		},
	}

	content := `
		:e3 2 |+7 |+4 |+3
	`

	logs.Debug("parse raw")
	irp, _ := bleeder.ParseContent(content, context)

	logs.Debug("new wav player")
	p := player.NewWAVPlayer(44100, 1)

	logs.Debug("play IR")
	p.Play(irp, 0, irp.Length())
}

/*
		@ chord5 e2 2

	~40_ |*2
	:e3 2_1 |+7 |+5 |-5

	_8
	:e2 4 |+7
	_4
	:g#2 4 |+7 _
	:c#2 4 |+7 _
	:a2 4 |+7 _

	_2
	>30 |+12 |+12 _


	>30 >42 >54

	:e2
	_/2
	:a3 d/2
	:e3 2 .9
	:f4 |/2
	:c3+2 4_+3 |+7
	:e4 _4 >60 4 _+4 | +2
*/
