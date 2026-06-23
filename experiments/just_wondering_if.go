package experiments

func run() {
	chord5 := new(Sequence).Note("e2", 1).Last("+7").Last("+5")

	new(Sequence).
		Vibe("synth").
		Volume(1.0).
		Note("e2", 1).Note("a2", 1).
		Wait(1).
		Volume(0.8).
		Note("e2", 1).Last("+7").Last("+5").
		Wait(1).
		Link(*chord5.Vars("d4"))
}

type Sequence struct{}

func (s *Sequence) Vars(args ...string) *Sequence {
	return s
}

// modificators
func (s *Sequence) Vibe(sample string) *Sequence {
	return s
}
func (s *Sequence) Volume(vol float64) *Sequence {
	return s
}

// commands
func (s *Sequence) Midi(midi float64, dur int) *Sequence {
	return s
}

func (s *Sequence) Note(note string, dur int) *Sequence {
	return s
}

func (s *Sequence) Wait(ticks int) *Sequence {
	return s
}

func (s *Sequence) Last(args ...string) *Sequence {
	return s
}

func (s *Sequence) Link(seq Sequence) *Sequence {
	return s
}
