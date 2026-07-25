package core

type BleedCtx struct {
	bleed   *Bleed
	bleeder *Bleeder
}

func NewBleedCtx(bleed *Bleed) *BleedCtx {
	return &BleedCtx{
		bleed:   bleed,
		bleeder: NewBleeder(bleed),
	}
}

func (b *BleedCtx) Play(name, vars string) error {
	return nil
}

func (ctx *BleedCtx) DoSome() {
	// TODO
}
