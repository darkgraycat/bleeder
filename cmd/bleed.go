package cmd

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

type Bleed struct {
	Meta struct {
		Main   string     `toml:"main"`
		Bleeds []BleedRef `toml:"bleeds"`
	} `toml:"meta"`
	Sequence map[string]struct {
		Args    Args    `toml:"args"`
		Repeat  int     `toml:"repeat"`
		Content Content `toml:"content"`
	} `toml:"seq"`
}

func LoadBleed(path string) (*Bleed, error) {
	var bleed Bleed
	if _, err := toml.DecodeFile(path, &bleed); err != nil {
		return nil, err
	}
	return &bleed, nil
}

type BleedRef struct {
	*Bleed
}

func (r *BleedRef) UnmarshalTOML(data any) error {
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

type Args map[string]string

func (a *Args) UnmarshalTOML(data any) error {
	s, ok := data.(string)
	if !ok {
		return fmt.Errorf("args should be string, got %T", data)
	}
	*a = make(Args)
	for part := range strings.FieldsSeq(s) {
		k, v, ok := strings.Cut(part, ":")
		if !ok {
			return fmt.Errorf("invalid arg: %q", part)
		}
		(*a)[k] = v
	}
	return nil
}

type Content []string

func (c *Content) UnmarshalTOML(data any) error {
	if s, ok := data.(string); ok {
		*c = strings.Fields(s)
		return nil
	}
	return fmt.Errorf("content should be whitespace character separated string, got %T", data)
}
