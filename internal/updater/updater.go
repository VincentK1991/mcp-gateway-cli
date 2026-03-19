// Package updater performs a non-blocking background check for newer releases
// on GitHub and prints a hint to stderr when one is available. It rate-limits
// itself to at most one check per 24 hours via a timestamp file.
package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	releaseAPIURL  = "https://api.github.com/repos/VincentK1991/mcp-gateway-cli/releases/latest"
	releasePageURL = "https://github.com/VincentK1991/mcp-gateway-cli/releases/latest"
	checkInterval  = 24 * time.Hour
	httpTimeout    = 5 * time.Second
)

// Check fetches the latest GitHub release tag and, if it is newer than
// current, prints a one-line hint to stderr. It is safe to call in a
// goroutine and silently no-ops on any error.
func Check(current string) {
	// Skip for local / CI builds.
	if current == "dev" || current == "" {
		return
	}

	if !dueForCheck() {
		return
	}

	latest, err := fetchLatestTag()
	if err != nil {
		return
	}

	_ = writeCheckTimestamp() // best-effort; ignore error

	if isNewer(latest, current) {
		fmt.Fprintf(os.Stderr,
			"\nA new version of gateway-cli is available: %s\n"+
				"  curl install : curl -fsSL https://raw.githubusercontent.com/VincentK1991/mcp-gateway-cli/main/install.sh | bash\n"+
				"  homebrew     : brew upgrade gateway-cli\n"+
				"  go install   : go install github.com/VincentK1991/mcp-gateway-cli@latest\n\n",
			latest,
		)
	}
}

// dueForCheck returns true if 24 hours have passed since the last check.
func dueForCheck() bool {
	ts, err := readCheckTimestamp()
	if err != nil {
		return true // no timestamp file → check now
	}
	return time.Since(ts) >= checkInterval
}

func timestampPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gateway-cli", "last-update-check"), nil
}

func readCheckTimestamp() (time.Time, error) {
	p, err := timestampPath()
	if err != nil {
		return time.Time{}, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, strings.TrimSpace(string(data)))
}

func writeCheckTimestamp() error {
	p, err := timestampPath()
	if err != nil {
		return err
	}
	return os.WriteFile(p, []byte(time.Now().UTC().Format(time.RFC3339)), 0644)
}

func fetchLatestTag() (string, error) {
	client := &http.Client{Timeout: httpTimeout}
	req, err := http.NewRequest(http.MethodGet, releaseAPIURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github API returned %d", resp.StatusCode)
	}

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	return payload.TagName, nil
}

// isNewer returns true when latest != current (simple string comparison;
// sufficient for semver tags like "v1.2.3" when the release pipeline always
// increments). A full semver comparison can be added later if needed.
func isNewer(latest, current string) bool {
	return latest != "" && latest != current
}
