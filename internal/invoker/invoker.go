package invoker

import (
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
	mcpclient "github.com/VincentK1991/mcp-gateway-cli/internal/mcp"
	"github.com/VincentK1991/mcp-gateway-cli/internal/schema"
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
	textOnly, _ := cmd.Root().PersistentFlags().GetBool("text")

	if jsonOnly && textOnly {
		return fmt.Errorf("--json and --text are mutually exclusive")
	}

	if textOnly {
		if len(result.Content) == 0 {
			return fmt.Errorf("tool returned no content")
		}
		var text string
		switch tc := result.Content[0].(type) {
		case mcp.TextContent:
			text = tc.Text
		case *mcp.TextContent:
			text = tc.Text
		default:
			return fmt.Errorf("first content item is not text (got %T)", result.Content[0])
		}
		fmt.Print(text)
		return nil
	}

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
