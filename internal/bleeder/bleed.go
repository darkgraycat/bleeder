package bleeder

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

// File-type parser
type Bleed struct {
	Meta      `toml:"meta"`
	Sequences map[string]Seq `toml:"seq"`
}

type Meta struct {
	Main    string     `toml:"main"`    // main sequence name
	Include []bleedRef `toml:"include"` // included bleeds
	Tempo   int        `toml:"tempo"`   // beats per minute
}

type Seq struct {
	Args    args   `toml:"args"`    // sequence arguments
	Repeat  int    `toml:"repeat"`  // repeats count
	Shape   string `toml:"shape"`   // shape of the wave (TODO)
	Content string `toml:"content"` // sequence contents
}

type bleedRef struct {
	*Bleed
}

type args []string

func LoadBleed(path string) (*Bleed, error) {
	var bleed Bleed
	if _, err := toml.DecodeFile(path, &bleed); err != nil {
		return nil, err
	}
	return &bleed, nil
}

func (b Bleed) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Main: %s\n", b.Meta.Main)
	if len(b.Meta.Include) > 0 {
		sb.WriteString("Includes:\n")
		for i, ref := range b.Meta.Include {
			fmt.Fprintf(&sb, "  [%d] %p\n", i, ref.Bleed)
		}
	}
	sb.WriteString("Sequences:\n")
	for k, v := range b.Sequences {
		fmt.Fprintf(&sb, "  %s: repeat=%d args=%v content=%v\n",
			k, v.Repeat, v.Args, v.Content)
	}
	return sb.String()
}

func (r *bleedRef) UnmarshalTOML(data any) error {
	if s, ok := data.(string); ok {
		bleed, err := LoadBleed(s)
		if err != nil {
			return err
		}
		r.Bleed = bleed
		return nil
	}
	return fmt.Errorf("bleeds should contain at filepath strings, got %T", data)
}

func (a *args) UnmarshalTOML(data any) error {
	s, ok := data.(string)
	if !ok {
		return fmt.Errorf("args should be string, got %T", data)
	}
	for part := range strings.FieldsSeq(s) {
		k, v, ok := strings.Cut(part, ":")
		if !ok {
			return fmt.Errorf("invalid arg: %q", part)
		}
		*a = append(*a, k, v)
	}
	return nil
}
