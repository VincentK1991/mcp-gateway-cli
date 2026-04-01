package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// MCPEntry holds connection config for a single MCP server.
// For HTTP servers set URL (and optionally Headers).
// For stdio servers set Command (and optionally Args and Env).
type MCPEntry struct {
	// HTTP transport
	URL     string            `mapstructure:"url"`
	Headers map[string]string `mapstructure:"headers"`

	// Stdio transport — Command is the executable, Args are its arguments,
	// Env are extra environment variables in KEY=VALUE format.
	Command string            `mapstructure:"command"`
	Args    []string          `mapstructure:"args"`
	Env     map[string]string `mapstructure:"env"`
}

// Config is the top-level config loaded from ~/.gateway-cli/config.yaml.
type Config struct {
	MCPs map[string]MCPEntry `mapstructure:"mcps"`
}

// Load reads the config file and returns the parsed Config.
// If cfgFile is empty, it defaults to ~/.gateway-cli/config.yaml.
func Load(cfgFile string) (*Config, error) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("could not determine home directory: %w", err)
		}
		cfgDir := filepath.Join(home, ".gateway-cli")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			return nil, fmt.Errorf("could not create config directory: %w", err)
		}
		viper.AddConfigPath(cfgDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	return &cfg, nil
}

// MCPURLs returns a flat map of MCP name → URL for convenience.
func (c *Config) MCPURLs() map[string]string {
	urls := make(map[string]string, len(c.MCPs))
	for name, entry := range c.MCPs {
		urls[name] = entry.URL
	}
	return urls
}
