package config

import (
	"errors"
	"os"

	"github.com/BurntSushi/toml"
)

func LoadGlobal(path string) (GlobalConfig, error) {
	var cfg GlobalConfig
	if _, err := os.Stat(path); err != nil {
		return GlobalConfig{}, err
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return GlobalConfig{}, err
	}
	if cfg.JDKs == nil {
		cfg.JDKs = map[string]string{}
	}
	if cfg.Mavens == nil {
		cfg.Mavens = map[string]string{}
	}
	return cfg, nil
}

func LoadProject(path string) (ProjectConfig, error) {
	var cfg ProjectConfig
	if _, err := os.Stat(path); err != nil {
		return ProjectConfig{}, err
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return ProjectConfig{}, err
	}
	return cfg, nil
}

func IsNotExist(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}
