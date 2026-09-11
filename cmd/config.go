package main

import (
	"os"
	"io/fs"
	"errors"
	"path/filepath"
	"fmt"

	toml "github.com/pelletier/go-toml/v2"
)

type Config struct {
	Shell string `toml:"shell"`
}

func defaultShell() string {
	if sh := os.Getenv("SHELL"); sh != "" {
		return sh
	}
	return "/bin/zsh"
}

func defaultConfig() Config {
	return Config{
		Shell: defaultShell(),
	}
}

func initConfig() (Config, error) {
	cfg := defaultConfig()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return cfg, fmt.Errorf("resolving user config dir: %w", err)
	}

	configPath := filepath.Join(homeDir, ".config", "shftty", "shftty.toml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return cfg, nil
		}

		return cfg, fmt.Errorf("reading config %s: %w", configPath, err)
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Shell == "" {
		cfg.Shell = defaultShell()
	}

	return cfg, nil
}
