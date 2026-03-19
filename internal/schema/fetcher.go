package schema

import (
	"fmt"
	"time"

	mcpclient "github.com/VincentK1991/mcp-gateway-cli/internal/mcp"
	mcptype "github.com/mark3labs/mcp-go/mcp"
)

// FetchAll queries each registered MCP server and returns the combined schema.
// Servers that fail are skipped with a warning — a partial result is still returned.
func FetchAll(endpoints map[string]MCPEndpoint) (*GatewaySchema, error) {
	gs := newGatewaySchema()
	gs.LastFetch = time.Now()

	for name, ep := range endpoints {
		tools, err := mcpclient.FetchTools(ep.URL, ep.Headers)
		if err != nil {
			fmt.Printf("Warning: skipping MCP '%s' (%s): %v\n", name, ep.URL, err)
			continue
		}

		entry := MCP{
			Name:  name,
			Tools: make(map[string]Tool, len(tools)),
		}
		for toolName, t := range tools {
			entry.Tools[toolName] = convertTool(t)
		}
		gs.MCPs[name] = entry
	}

	return gs, nil
}

// convertTool maps an mcp-go Tool to our internal Tool type.
func convertTool(t mcptype.Tool) Tool {
	props := make(map[string]Property, len(t.InputSchema.Properties))
	for name, raw := range t.InputSchema.Properties {
		if m, ok := raw.(map[string]interface{}); ok {
			propType, _ := m["type"].(string)
			propDesc, _ := m["description"].(string)
			props[name] = Property{Type: propType, Description: propDesc}
		}
	}

	return Tool{
		Name:        t.Name,
		Description: t.Description,
		InputSchema: InputSchema{
			Type:       t.InputSchema.Type,
			Properties: props,
			Required:   t.InputSchema.Required,
		},
	}
}
