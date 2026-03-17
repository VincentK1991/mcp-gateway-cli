package invoker

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	mcpclient "mcp-gateway-cli/internal/mcp"
	"mcp-gateway-cli/internal/schema"
)

// BuildToolCommand creates a Cobra subcommand for a specific MCP tool.
// Flags are generated dynamically from the tool's input schema.
func BuildToolCommand(mcpName, toolName string, tool schema.Tool, endpoint schema.MCPEndpoint) *cobra.Command {
	cmd := &cobra.Command{
		Use:   toolName,
		Short: tool.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTool(cmd, endpoint, toolName, tool)
		},
	}

	required := make(map[string]bool, len(tool.InputSchema.Required))
	for _, r := range tool.InputSchema.Required {
		required[r] = true
	}

	for name, prop := range tool.InputSchema.Properties {
		cmd.Flags().String(name, "", prop.Description)
		if required[name] {
			cmd.MarkFlagRequired(name)
		}
	}

	return cmd
}

func runTool(cmd *cobra.Command, endpoint schema.MCPEndpoint, toolName string, tool schema.Tool) error {
	params := make(map[string]interface{}, len(tool.InputSchema.Properties))
	for name := range tool.InputSchema.Properties {
		if val, err := cmd.Flags().GetString(name); err == nil && val != "" {
			params[name] = val
		}
	}

	result, err := mcpclient.CallTool(endpoint.URL, toolName, params, endpoint.Headers)
	if err != nil {
		return err
	}

	jsonOnly, _ := cmd.Root().PersistentFlags().GetBool("json")

	var payload any
	if jsonOnly {
		payload = result.StructuredContent
	} else {
		payload = result
	}

	out, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode result: %w", err)
	}

	fmt.Println(string(out))
	return nil
}
