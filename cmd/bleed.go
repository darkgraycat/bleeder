package cmd

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

type Bleed struct {
	Meta struct {
		Main   string     `toml:"main"`
		Bleeds []bleedRef `toml:"bleeds"`
	} `toml:"meta"`
	Sequences map[string]struct {
		Args    args    `toml:"args"`
		Repeat  int     `toml:"repeat"`
		Shape   string  `toml:"shape"` // TODO: use it
		Content content `toml:"content"`
	} `toml:"seq"`
}

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
	if len(b.Meta.Bleeds) > 0 {
		sb.WriteString("Includes:\n")
		for i, ref := range b.Meta.Bleeds {
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

type bleedRef struct {
	*Bleed
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

type args map[string]string

func (a *args) UnmarshalTOML(data any) error {
	s, ok := data.(string)
	if !ok {
		return fmt.Errorf("args should be string, got %T", data)
	}
	*a = make(args)
	for part := range strings.FieldsSeq(s) {
		k, v, ok := strings.Cut(part, ":")
		if !ok {
			return fmt.Errorf("invalid arg: %q", part)
		}
		(*a)[k] = v
	}
	return nil
}

type content []string

func (c *content) UnmarshalTOML(data any) error {
	if s, ok := data.(string); ok {
		*c = strings.Split(s, "\n")
		return nil
	}
	return fmt.Errorf("content should be whitespace character separated string, got %T", data)
}
