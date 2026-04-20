package cmd

import "github.com/BurntSushi/toml"

type Config struct {
	Audio struct {
		SampleRate int `toml:"sample_rate"`
		Channels   int `toml:"channels"`
	} `toml:"audio"`

	Parser struct {
		DefaultDur float64 `toml:"default_dur"`
		DefaultVol float64 `toml:"default_vol"`
	} `toml:"parser"`

	Output struct {
		Command string
		Args    []string
	} `toml:"output"`

	Mapping struct {
		Play       string `toml:"play"`
		Wave       string `toml:"wave"`
		Seq        string `toml:"seq"`
		Wait       string `toml:"wait"`
		Repeat     string `toml:"repeat"`
		RepeatLine string `toml:"repeat_line"`
		Debug      string `toml:"debug"`
	} `toml:"mapping"`
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
