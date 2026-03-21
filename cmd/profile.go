package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/VincentK1991/mcp-gateway-cli/internal/config"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage gateway-cli profiles",
	Long: `Manage named profiles. Each profile holds its own set of MCP servers.

Example config:

  current-profile: dev

  profiles:
    dev:
      mcps:
        calculator:
          url: http://localhost:8000/mcp
    production:
      mcps:
        calculator:
          url: https://prod.example.com/mcp
          headers:
            Authorization: "Bearer ${PROD_TOKEN}"`,
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.IsLegacy() {
			fmt.Println("* default (legacy flat config — no profiles defined)")
			return nil
		}

		names := cfg.ProfileNames()
		if len(names) == 0 {
			fmt.Println("No profiles defined. Add a 'profiles:' section to your config.")
			return nil
		}

		active := cfg.ActiveProfile(profileOverride)
		for _, name := range names {
			if name == active {
				fmt.Printf("* %s\n", name)
			} else {
				fmt.Printf("  %s\n", name)
			}
		}
		return nil
	},
}

var profileShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the currently active profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		active := cfg.ActiveProfile(profileOverride)
		if active == "" {
			return fmt.Errorf("no active profile set; run 'gateway-cli profile use <name>'")
		}
		if cfg.IsLegacy() {
			fmt.Printf("%s (legacy flat config)\n", active)
			return nil
		}
		if profileOverride != "" {
			fmt.Printf("%s (temporary override via --profile)\n", active)
		} else {
			fmt.Println(active)
		}
		return nil
	},
}

var profileUseCmd = &cobra.Command{
	Use:   "use <profile>",
	Short: "Switch the active profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := config.SetCurrentProfile(cfgFile, name, cfg); err != nil {
			return err
		}
		fmt.Printf("Switched to profile %q.\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(profileCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileShowCmd)
	profileCmd.AddCommand(profileUseCmd)
}
