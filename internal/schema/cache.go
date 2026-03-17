package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	cacheFileName = "schema-cache.json"
	DefaultTTL    = 1 * time.Hour
)

func cachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gateway-cli", cacheFileName), nil
}

// LoadCache reads the cached GatewaySchema from disk.
func LoadCache() (*GatewaySchema, error) {
	path, err := cachePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no cache found")
		}
		return nil, err
	}

	var gs GatewaySchema
	if err := json.Unmarshal(data, &gs); err != nil {
		return nil, fmt.Errorf("corrupted cache: %w", err)
	}
	return &gs, nil
}

// SaveCache writes the GatewaySchema to disk.
func SaveCache(gs *GatewaySchema) error {
	path, err := cachePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(gs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// IsStale returns true if the cache is missing or older than ttl.
func IsStale(gs *GatewaySchema, ttl time.Duration) bool {
	return gs == nil || gs.LastFetch.IsZero() || time.Since(gs.LastFetch) > ttl
}

// InvalidateCache deletes the cache file.
func InvalidateCache() error {
	path, err := cachePath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
