package invoker

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
	mcpclient "github.com/VincentK1991/mcp-gateway-cli/internal/mcp"
	"github.com/VincentK1991/mcp-gateway-cli/internal/schema"
)

// maxOutputChars is the threshold above which output is written to a file
// instead of stdout. 40,000 chars ≈ 10,000 tokens at ~4 chars/token.
const maxOutputChars = 40_000

// isTerminal reports whether stdout is an interactive terminal (not a pipe or redirect).
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// writeOutputToFile writes data to a timestamped file in the current directory
// and returns the filename.
func writeOutputToFile(data []byte, ext string) (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("gateway-output-%s.%s", timestamp, ext)
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return "", err
	}
	return filename, nil
}

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

	// Write to file only when: output is oversized and stdout is an interactive
	// terminal. If stdout is a pipe or redirect, another program is consuming
	// the output, so always pass it through unchanged.
	if isTerminal() && len(out) > maxOutputChars {
		filename, writeErr := writeOutputToFile(out, "json")
		if writeErr == nil {
			fmt.Fprintf(os.Stderr, "Output too large (%d KB, ~%d tokens) — written to: %s\n",
				len(out)/1024, len(out)/4, filename)
			return nil
		}
		// Fall back to stdout if the file write fails.
		fmt.Fprintf(os.Stderr, "Warning: could not write output file (%v); printing to stdout\n", writeErr)
	}

	fmt.Println(string(out))
	return nil
}
