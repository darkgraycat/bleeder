package core

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Bleed entrypoint name
const MAIN_NAME = "main"

// Bleed is the top-level structure representing a parsed .bleed file.
type Bleed struct {
	Meta      Meta                `toml:"meta"`  // metadata
	Patches   map[string]Patch    `toml:"patch"` // sound data
	Sequences map[string]Sequence `toml:"seq"`   // sequence data
}

// Meta holds global playback settings for a bleed file.
type Meta struct {
	Path    string   // file path
	Include []string `toml:"include"` // included bleed file paths
}

// Audio modification
type Patch struct {
	Wave string // name of wave function to use
}

// Sequence defines a named playback data using DSL
type Sequence struct {
	Vars    string `toml:"vars"`    // sequence arguments
	Tick    string `toml:"tick"`    // sequence tick duration
	Tune    string `toml:"tune"`    // sequence transposition
	Gain    string `toml:"gain"`    // sequence volume
	Content string `toml:"content"` // sequence content
}

// Load Bleed file contents
func LoadBleed(path string) (*Bleed, error) {
	b := &Bleed{
		Meta:      Meta{Path: path},
		Patches:   make(map[string]Patch),
		Sequences: make(map[string]Sequence),
	}
	if _, err := toml.DecodeFile(path, &b); err != nil {
		return nil, err
	}
	// validate namings
	for k := range b.Sequences {
		if strings.ContainsAny(k, "+-*/%^@$_") {
			return nil, fmt.Errorf("sequence %q name invalid", k)
		}
	}
	for k := range b.Patches {
		if strings.ContainsAny(k, "+-*/%^@$_") {
			return nil, fmt.Errorf("patch %q name invalid", k)
		}
	}
	// parse included bleeds
	baseDir := filepath.Dir(path)
	for _, includePath := range b.Meta.Include {
		included, err := LoadBleed(filepath.Join(baseDir, includePath))
		if err != nil {
			return nil, err
		}
		// load patches
		for k, v := range included.Patches {
			log.Printf("[INIT] load patch %q from %q\n", k, includePath)
			if _, exists := b.Patches[k]; exists {
				return nil, fmt.Errorf("patch %q already exists, conflict with include %q", k, includePath)
			}
			b.Patches[k] = v
		}
		// load sequences
		for k, v := range included.Sequences {
			log.Printf("[INIT] load sequence %q from %q\n", k, includePath)
			if _, exist := b.Sequences[k]; exist {
				return nil, fmt.Errorf("sequence %q already exists, conflict with include %q", k, includePath)
			}
			b.Sequences[k] = v
		}
	}
	return b, nil
}

func (b Bleed) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n", b.Meta)

	sb.WriteString("Sequences:\n")
	for k := range b.Sequences {
		fmt.Fprintf(&sb, "  - %s\n", k)
	}
	return sb.String()
}

func (s Sequence) String() string {
	return fmt.Sprintf("args=%q content=%q", s.Vars, s.Content)
}

func (m Meta) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Bleed %q\n", m.Path)
	if len(m.Include) > 0 {
		sb.WriteString("Includes:\n")
		for _, path := range m.Include {
			fmt.Fprintf(&sb, "  - %s\n", path)
		}
	}
	return sb.String()
}
