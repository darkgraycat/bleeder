package cmd

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Audio struct {
		SampleRate int `toml:"sample_rate"`
		Channels   int `toml:"channels"`
	} `toml:"audio"`

	Live struct {
		Port int `toml:"port"`
	}
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func defaultConfigPath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	dir := filepath.Dir(ex)
	return filepath.Join(dir, "config.toml")
}
