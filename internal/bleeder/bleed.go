package bleeder

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// SequenceType describes how a sequence content should be parsed
type SequenceType int

// Bleed entrypoint name
const MAIN_NAME = "main"

// Sequence types enum
const (
	SEQ_UNKNOWN SequenceType = iota
	SEQ_LANE
	SEQ_RIFF
)

// Bleed is the top-level structure representing a parsed .bleed file.
type Bleed struct {
	Meta  Meta                `toml:"meta"` // metadata
	Vibes map[string]Vibe     `toml:"vibe"` // named vibes
	Lanes map[string]Sequence `toml:"lane"` // named lanes
	Riffs map[string]Sequence `toml:"riff"` // named riffs
}

// Meta holds global playback settings for a bleed file.
type Meta struct {
	Path    string   // file path
	Tempo   int      `toml:"tempo"`   // beats per minute
	Include []string `toml:"include"` // included bleed file paths
}

// Audio modification
type Vibe struct {
	Wave string // name of wave function to use
}

// Sequence defines a named playback data using DSL
type Sequence struct {
	Type    SequenceType
	Vars    string `toml:"vars"`    // sequence arguments
	Content string `toml:"content"` // sequence content
}

// Load Bleed file contents
func LoadBleed(path string) (*Bleed, error) {
	b := &Bleed{
		Meta:  Meta{Path: path},
		Vibes: make(map[string]Vibe),
		Lanes: make(map[string]Sequence),
		Riffs: make(map[string]Sequence),
	}
	if _, err := toml.DecodeFile(path, &b); err != nil {
		return nil, err
	}
	// assign lane type and validate naming
	for k, v := range b.Lanes {
		if strings.ContainsAny(k, chRest+"+-*/") {
			return nil, fmt.Errorf("sequence %q name cannot contain special characters", k)
		}
		if _, exists := b.Riffs[k]; exists {
			return nil, fmt.Errorf("sequence %q defined in both lane and riff", k)
		}
		v.Type = SEQ_LANE
		b.Lanes[k] = v
	}
	// assign riff type and validate naming
	for k, v := range b.Riffs {
		if strings.ContainsAny(k, chRest) {
			return nil, fmt.Errorf("sequence %q name cannot contain %s", k, chRest)
		}
		if _, exists := b.Lanes[k]; exists {
			return nil, fmt.Errorf("sequence %q defined in both lane and riff", k)
		}
		v.Type = SEQ_RIFF
		b.Riffs[k] = v
	}
	// validate vibe naming
	for k := range b.Vibes {
		if strings.ContainsAny(k, chRest) {
			return nil, fmt.Errorf("vibe %q name cannot contain %s", k, chRest)
		}
	}
	// parse included bleeds
	baseDir := filepath.Dir(path)
	for _, includePath := range b.Meta.Include {
		included, err := LoadBleed(filepath.Join(baseDir, includePath))
		if err != nil {
			return nil, err
		}
		// load vibes
		for k, v := range included.Vibes {
			fmt.Printf("Load vibe %s from %s\n", k, includePath)
			if _, exists := b.Vibes[k]; exists {
				return nil, fmt.Errorf("vibe %q already defined, conflict with include %q", k, includePath)
			}
			b.Vibes[k] = v
		}
		// load lanes
		for k, v := range included.Lanes {
			fmt.Printf("Load lane %s from %s\n", k, includePath)
			if _, exists := b.Lanes[k]; exists {
				return nil, fmt.Errorf("lane %q already defined, conflict with include %q", k, includePath)
			}
			b.Lanes[k] = v
		}
		// load riffs
		for k, v := range included.Riffs {
			fmt.Printf("Load riff %s from %s\n", k, includePath)
			if _, exists := b.Riffs[k]; exists {
				return nil, fmt.Errorf("riff %q already defined, conflict with include %q", k, includePath)
			}
			b.Riffs[k] = v
		}
	}
	return b, nil
}

func (b Bleed) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n", b.Meta)

	sb.WriteString("Lanes:\n")
	for k := range b.Lanes {
		fmt.Fprintf(&sb, "  - %s\n", k)
	}
	sb.WriteString("Riffs:\n")
	for k := range b.Riffs {
		fmt.Fprintf(&sb, "  - %s\n", k)
	}
	return sb.String()
}

func (s Sequence) String() string {
	return fmt.Sprintf("args=%q content=%q", s.Vars, s.Content)
}

func (m Meta) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Bleed: %s\n", m.Path)
	if len(m.Include) > 0 {
		sb.WriteString("Includes:\n")
		for _, path := range m.Include {
			fmt.Fprintf(&sb, "  - %s\n", path)
		}
	}
	return sb.String()
}
