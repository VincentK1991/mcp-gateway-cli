package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/VincentK1991/mcp-gateway-cli/internal/schema"
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Manage the local MCP schema cache",
}

var schemaRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Fetch schemas from all registered MCP servers and update the cache",
	RunE: func(cmd *cobra.Command, args []string) error {
		endpoints := mcpEndpoints()
		fmt.Printf("Fetching schemas from %d MCP server(s)...\n", len(endpoints))

		fetched, err := schema.FetchAll(endpoints)
		if err != nil {
			return fmt.Errorf("fetch failed: %w", err)
		}

		activeProfile := cfg.ActiveProfile(profileOverride)
		if err := schema.SaveCache(fetched, activeProfile); err != nil {
			return fmt.Errorf("failed to save cache: %w", err)
		}

		fmt.Println("Schema cache updated.")
		return nil
	},
}

var schemaInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show details about the currently cached schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		activeProfile := cfg.ActiveProfile(profileOverride)
		cached, err := schema.LoadCache(activeProfile)
		if err != nil {
			return fmt.Errorf("no cache available: %w", err)
		}

		toolCount := 0
		for _, mcp := range cached.MCPs {
			toolCount += len(mcp.Tools)
		}

		fmt.Printf("Profile      : %s\n", activeProfile)
		fmt.Printf("Last fetched : %s\n", cached.LastFetch.Format("2006-01-02 15:04:05"))
		fmt.Printf("MCP servers  : %d\n", len(cached.MCPs))
		fmt.Printf("Total tools  : %d\n", toolCount)

		for name, mcp := range cached.MCPs {
			fmt.Printf("\n  %s (%d tools)\n", name, len(mcp.Tools))
			for toolName, tool := range mcp.Tools {
				fmt.Printf("    %-20s %s\n", toolName, tool.Description)
			}
		}
		return nil
	},
}

var schemaInvalidateCmd = &cobra.Command{
	Use:   "invalidate",
	Short: "Delete the local schema cache for the active profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		activeProfile := cfg.ActiveProfile(profileOverride)
		if err := schema.InvalidateCache(activeProfile); err != nil {
			return fmt.Errorf("failed to invalidate cache: %w", err)
		}
		fmt.Printf("Cache invalidated for profile %q.\n", activeProfile)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(schemaCmd)
	schemaCmd.AddCommand(schemaRefreshCmd)
	schemaCmd.AddCommand(schemaInfoCmd)
	schemaCmd.AddCommand(schemaInvalidateCmd)
}
