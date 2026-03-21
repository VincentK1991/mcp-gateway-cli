package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/VincentK1991/mcp-gateway-cli/internal/config"
	"github.com/VincentK1991/mcp-gateway-cli/internal/invoker"
	"github.com/VincentK1991/mcp-gateway-cli/internal/schema"
	"github.com/VincentK1991/mcp-gateway-cli/internal/updater"
)

// mcpEndpoints converts the loaded config into a map of schema.MCPEndpoint,
// expanding any environment variable references in header values.
func mcpEndpoints() map[string]schema.MCPEndpoint {
	mcps, err := cfg.ActiveMCPs(profileOverride)
	if err != nil {
		return nil
	}
	endpoints := make(map[string]schema.MCPEndpoint, len(mcps))
	for name, entry := range mcps {
		ep := schema.MCPEndpoint{URL: entry.URL}
		if len(entry.Headers) > 0 {
			ep.Headers = make(map[string]string, len(entry.Headers))
			for k, v := range entry.Headers {
				ep.Headers[k] = os.ExpandEnv(v)
			}
		}
		endpoints[name] = ep
	}
	return endpoints
}

var (
	cfgFile        string
	profileOverride string
	refreshSchema  bool
	offline        bool
	cfg            *config.Config
)

var rootCmd = &cobra.Command{
	Use:     "gateway-cli",
	Short:   "A CLI to interact with MCP servers",
	Long:    `gateway-cli fetches and caches tool schemas from configured MCP servers and exposes them as CLI commands.`,
	Version: Version,
}

// Execute is the entrypoint called from main.
func Execute() {
	var err error
	cfg, err = config.Load(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	go updater.Check(Version)

	// Build dynamic tool commands before Cobra routes the command.
	if err := buildToolCommands(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing tool commands: %v\n", err)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.gateway-cli/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&profileOverride, "profile", "", "Use a specific profile for this command (overrides current-profile)")
	rootCmd.PersistentFlags().BoolVar(&refreshSchema, "refresh-schema", false, "Force re-fetch schemas from all MCP servers, ignoring cache TTL")
	rootCmd.PersistentFlags().BoolVar(&offline, "offline", false, "Use cached schema only, never contact MCP servers")
	rootCmd.PersistentFlags().BoolP("json", "j", false, "Output only the structured content (useful for piping)")
	rootCmd.PersistentFlags().BoolP("text", "t", false, "Output only the first text content item as a plain string (useful for piping)")
}

// buildToolCommands resolves the active schema and registers one subcommand
// per MCP server and tool discovered.
func buildToolCommands() error {
	// Skip schema loading for built-in management commands.
	if len(os.Args) > 1 && (os.Args[1] == "schema" || os.Args[1] == "help" || os.Args[1] == "profile") {
		return nil
	}

	// Parse flags manually so they're available before Cobra routes.
	args := os.Args[1:]
	for i, a := range args {
		switch a {
		case "--refresh-schema":
			refreshSchema = true
		case "--offline":
			offline = true
		case "--profile":
			if i+1 < len(args) {
				profileOverride = args[i+1]
			}
		}
	}

	activeSchema, err := resolveSchema()
	if err != nil {
		return err
	}

	endpoints := mcpEndpoints()

	for mcpName, mcp := range activeSchema.MCPs {
		ep := endpoints[mcpName]
		mcpCmd := &cobra.Command{
			Use:   mcpName,
			Short: fmt.Sprintf("Tools from the '%s' MCP server", mcpName),
		}
		for toolName, tool := range mcp.Tools {
			mcpCmd.AddCommand(invoker.BuildToolCommand(mcpName, toolName, tool, ep))
		}
		rootCmd.AddCommand(mcpCmd)
	}

	return nil
}

// resolveSchema returns a valid GatewaySchema either from cache or by fetching.
func resolveSchema() (*schema.GatewaySchema, error) {
	activeProfile := cfg.ActiveProfile(profileOverride)
	if activeProfile == "" {
		return nil, fmt.Errorf("no active profile set; run 'gateway-cli profile use <name>'")
	}

	cached, cacheErr := schema.LoadCache(activeProfile)
	cacheOK := cacheErr == nil && cached != nil

	needsFetch := refreshSchema || !cacheOK || schema.IsStale(cached, schema.DefaultTTL)

	if offline {
		if !cacheOK {
			return nil, fmt.Errorf("--offline requested but no cache found; run 'schema refresh' first")
		}
		return cached, nil
	}

	if !needsFetch {
		return cached, nil
	}

	fetched, err := schema.FetchAll(mcpEndpoints())
	if err != nil {
		return nil, err
	}

	if saveErr := schema.SaveCache(fetched, activeProfile); saveErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save schema cache: %v\n", saveErr)
	}

	if len(fetched.MCPs) == 0 && cacheOK {
		fmt.Fprintf(os.Stderr, "Warning: all MCP servers failed; falling back to cached schema from %s\n",
			cached.LastFetch.Format(time.RFC3339))
		return cached, nil
	}

	return fetched, nil
}
