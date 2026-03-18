package cmd

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

type Bleed struct {
	Include map[string]struct {
		Path string `toml:"path"`
	} `toml:"include"`
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

type Args []string

func (a *Args) UnmarshalTOML(data any) error {
	if s, ok := data.(string); ok {
		*a = strings.Fields(s)
		return nil
	}
	return fmt.Errorf("args should be whitespace character separated string, got %T", data)
}

type Content []string

func (c *Content) UnmarshalTOML(data any) error {
	if s, ok := data.(string); ok {
		*c = strings.Fields(s)
		return nil
	}
	return fmt.Errorf("content should be whitespace character separated string, got %T", data)
}
