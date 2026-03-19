package cmd

import "github.com/BurntSushi/toml"

type Config struct {
	Audio struct {
		SampleRate int `toml:"sample_rate"`
		Channels   int `toml:"channels"`
	} `toml:"audio"`

	Output struct {
		Command string
		Args    []string
	} `toml:"output"`

	Commands struct {
		Play       string `toml:"play"`
		Wait       string `toml:"wait"`
		Repeat     string `toml:"repeat"`
		RepeatLine string `toml:"repeat_line"`
	} `toml:"commands"`

	Symbols struct {
		Reference string `toml:"reference"`
	} `toml:"symbols"`
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
