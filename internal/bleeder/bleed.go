package bleeder

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

// SequenceType describes how a sequence content should be parsed
type SequenceType int

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
	Main    string   `toml:"main"`    // main sequence name
	Tempo   int      `toml:"tempo"`   // beats per minute
	Include []string `toml:"include"` // included bleed file paths
}

// Audio modification TODO:
type Vibe struct {
}

// Sequence defines a named playback data using DSL
type Sequence struct {
	Type    SequenceType
	Vars    string `toml:"vars"`    // sequence arguments
	Content string `toml:"content"` // sequence content
}

func LoadBleed(path string) (*Bleed, error) {
	var bleed Bleed
	if _, err := toml.DecodeFile(path, &bleed); err != nil {
		return nil, err
	}
	// assign lane type and validate naming
	for k, v := range bleed.Lanes {
		if strings.ContainsAny(k, chRest) {
			return nil, fmt.Errorf("sequence %q name cannot contain %s", k, chRest)
		}
		if _, exists := bleed.Riffs[k]; exists {
			return nil, fmt.Errorf("sequence %q defined in both lane and riff", k)
		}
		v.Type = SEQ_LANE
		bleed.Lanes[k] = v
	}
	// assign riff type and validate naming
	for k, v := range bleed.Riffs {
		if strings.ContainsAny(k, chRest) {
			return nil, fmt.Errorf("sequence %q name cannot contain %s", k, chRest)
		}
		if _, exists := bleed.Lanes[k]; exists {
			return nil, fmt.Errorf("sequence %q defined in both lane and riff", k)
		}
		v.Type = SEQ_RIFF
		bleed.Riffs[k] = v
	}
	// validate vibe naming
	for k := range bleed.Vibes {
		if strings.ContainsAny(k, chRest) {
			return nil, fmt.Errorf("vibe %q name cannot contain %s", k, chRest)
		}
	}
	// parse included bleeds
	for _, includePath := range bleed.Meta.Include {
		included, err := LoadBleed(includePath)
		if err != nil {
			return nil, err
		}
		// load vibes
		for k, v := range included.Vibes {
			if _, exists := bleed.Vibes[k]; exists {
				return nil, fmt.Errorf("vibe %q already defined, conflict with include %q", k, includePath)
			}
			bleed.Vibes[k] = v
		}
		// load lanes
		for k, v := range included.Lanes {
			if _, exists := bleed.Lanes[k]; exists {
				return nil, fmt.Errorf("lane %q already defined, conflict with include %q", k, includePath)
			}
			bleed.Lanes[k] = v
		}
		// load riffs
		for k, v := range included.Riffs {
			if _, exists := bleed.Riffs[k]; exists {
				return nil, fmt.Errorf("riff %q already defined, conflict with include %q", k, includePath)
			}
			bleed.Riffs[k] = v
		}
	}
	return &bleed, nil
}

func (b Bleed) String() string {
	var sb strings.Builder
	sb.WriteString("Meta:\n")
	fmt.Fprintf(&sb, "%s\n", b.Meta)

	sb.WriteString("Lanes:\n")
	for k, v := range b.Lanes {
		fmt.Fprintf(&sb, "  %s: %s\n", k, v)
	}
	sb.WriteString("Riffs:\n")
	for k, v := range b.Riffs {
		fmt.Fprintf(&sb, "  %s: %s\n", k, v)
	}
	return sb.String()
}

func (s Sequence) String() string {
	return fmt.Sprintf("args=%q content=%q", s.Vars, s.Content)
}

func (m Meta) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Main: %s\n", m.Main)
	if len(m.Include) > 0 {
		sb.WriteString("Includes:\n")
		for _, path := range m.Include {
			fmt.Fprintf(&sb, "  %s\n", path)
		}
	}
	return sb.String()
}
