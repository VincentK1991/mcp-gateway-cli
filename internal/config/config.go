package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// MCPEntry holds connection config for a single MCP server.
type MCPEntry struct {
	URL     string            `mapstructure:"url"  yaml:"url"`
	Headers map[string]string `mapstructure:"headers" yaml:"headers,omitempty"`
}

// ProfileEntry holds MCPs for a single named profile.
type ProfileEntry struct {
	MCPs map[string]MCPEntry `mapstructure:"mcps" yaml:"mcps"`
}

// Config is the top-level config loaded from ~/.gateway-cli/config.yaml.
type Config struct {
	// Legacy flat format: top-level mcps key (old configs).
	MCPs map[string]MCPEntry `mapstructure:"mcps" yaml:"mcps,omitempty"`

	// Profile format.
	CurrentProfile string                  `mapstructure:"current-profile" yaml:"current-profile,omitempty"`
	Profiles       map[string]ProfileEntry `mapstructure:"profiles"        yaml:"profiles,omitempty"`
}

// IsLegacy returns true when the config uses the old flat mcps format
// (top-level mcps key with no profiles block).
func (c *Config) IsLegacy() bool {
	return len(c.MCPs) > 0 && len(c.Profiles) == 0
}

// ActiveProfile returns the name of the currently active profile.
// The override (from --profile flag) takes precedence, then current-profile,
// then "default" for legacy flat configs.
func (c *Config) ActiveProfile(override string) string {
	if c.IsLegacy() {
		return "default"
	}
	if override != "" {
		return override
	}
	if c.CurrentProfile != "" {
		return c.CurrentProfile
	}
	return ""
}

// ActiveMCPs returns the MCPEntry map for the resolved active profile.
// It handles both the legacy flat format and the new profiles format.
func (c *Config) ActiveMCPs(override string) (map[string]MCPEntry, error) {
	if c.IsLegacy() {
		return c.MCPs, nil
	}

	profileName := override
	if profileName == "" {
		profileName = c.CurrentProfile
	}
	if profileName == "" {
		return nil, fmt.Errorf("no active profile set; run 'gateway-cli profile use <name>'")
	}

	p, ok := c.Profiles[profileName]
	if !ok {
		return nil, fmt.Errorf("profile %q not found in config", profileName)
	}
	return p.MCPs, nil
}

// ProfileNames returns a sorted list of all defined profile names.
func (c *Config) ProfileNames() []string {
	names := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
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

// SetCurrentProfile updates the current-profile field in the config file on disk.
// Returns an error if the config is in legacy flat format or if the profile doesn't exist.
func SetCurrentProfile(cfgFile string, profileName string, cfg *Config) error {
	if cfg.IsLegacy() {
		return fmt.Errorf("config uses legacy flat format; convert to profiles format before switching profiles")
	}
	if _, ok := cfg.Profiles[profileName]; !ok {
		return fmt.Errorf("profile %q not found in config", profileName)
	}

	path := cfgFile
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not determine home directory: %w", err)
		}
		path = filepath.Join(home, ".gateway-cli", "config.yaml")
	}

	// Read raw YAML as generic map to preserve unknown keys.
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading config: %w", err)
	}

	var raw map[string]interface{}
	if len(data) > 0 {
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing config: %w", err)
		}
	}
	if raw == nil {
		raw = make(map[string]interface{})
	}

	raw["current-profile"] = profileName

	out, err := yaml.Marshal(raw)
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	// Atomic write via temp file + rename.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("saving config: %w", err)
	}
	return nil
}
