package cmd

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

type Bleed struct {
	Include map[string]struct {
		Path string `toml:"include"`
	} `toml:"include"`

	Sequence map[string]struct {
		Args    Args   `toml:"args"`
		Content string `toml:"content"`
		Repeat  int    `toml:"repear"`
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
	switch v := data.(type) {
	case string:
		*a = strings.Split(v, " ")
		return nil
	default:
		return fmt.Errorf("command key must be a string or int, got %T", data)
	}
}
