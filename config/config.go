package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Addr            string `toml:"addr"`
	DSN             string `toml:"dsn"`
	IndexToken      string `toml:"index_token"`
	SnippetMaxRunes int    `toml:"snippet_max_runes"`
	Debug           bool   `toml:"debug"`
}

func Load(path string) (Config, error) {
	if strings.TrimSpace(path) == "" {
		return Config{}, fmt.Errorf("config path is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse toml config: %w", err)
	}

	if strings.TrimSpace(cfg.Addr) == "" {
		return Config{}, fmt.Errorf("addr is required")
	}
	if strings.TrimSpace(cfg.DSN) == "" {
		return Config{}, fmt.Errorf("dsn is required")
	}
	if strings.TrimSpace(cfg.IndexToken) == "" {
		return Config{}, fmt.Errorf("index_token is required")
	}
	if cfg.SnippetMaxRunes <= 0 {
		return Config{}, fmt.Errorf("snippet_max_runes must be greater than 0")
	}

	return cfg, nil
}
