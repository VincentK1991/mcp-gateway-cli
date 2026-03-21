package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

const DefaultTTL = 1 * time.Hour

var unsafeChars = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

// sanitizeProfileName replaces characters that are unsafe in filenames with "-".
func sanitizeProfileName(p string) string {
	return unsafeChars.ReplaceAllString(p, "-")
}

func cachePath(profile string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	name := fmt.Sprintf("schema-cache-%s.json", sanitizeProfileName(profile))
	return filepath.Join(home, ".gateway-cli", name), nil
}

// LoadCache reads the cached GatewaySchema from disk for the given profile.
func LoadCache(profile string) (*GatewaySchema, error) {
	path, err := cachePath(profile)
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

// SaveCache writes the GatewaySchema to disk for the given profile.
func SaveCache(gs *GatewaySchema, profile string) error {
	path, err := cachePath(profile)
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

// InvalidateCache deletes the cache file for the given profile.
func InvalidateCache(profile string) error {
	path, err := cachePath(profile)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
