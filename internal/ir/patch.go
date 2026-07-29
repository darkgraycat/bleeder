package ir

import (
	"bleeder/internal/audio"
	"fmt"
)

// Instruction shape of the sound
type Patch struct {
	Name     string         // patch name
	WaveFunc audio.WaveFunc // wave function to use
}

// Format Instruction into string
func (p *Patch) String() string {
	return fmt.Sprintf("Name=%s", p.Name)
}
