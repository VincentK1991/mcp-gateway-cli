package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	clientName    = "gateway-cli"
	clientVersion = "1.0.0"
)

// connect creates and initializes a streamable HTTP client for the given URL.
func connect(ctx context.Context, url string) (*client.Client, error) {
	c, err := client.NewStreamableHttpClient(url)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for %s: %w", url, err)
	}

	if err := c.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start client: %w", err)
	}

	req := mcp.InitializeRequest{}
	req.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	req.Params.ClientInfo = mcp.Implementation{Name: clientName, Version: clientVersion}

	if _, err := c.Initialize(ctx, req); err != nil {
		c.Close()
		return nil, fmt.Errorf("failed to initialize MCP session: %w", err)
	}

	return c, nil
}

// FetchTools connects to an MCP server and returns its available tools.
func FetchTools(url string) (map[string]mcp.Tool, error) {
	ctx := context.Background()

	c, err := connect(ctx, url)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	res, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	tools := make(map[string]mcp.Tool, len(res.Tools))
	for _, t := range res.Tools {
		tools[t.Name] = t
	}
	return tools, nil
}

// CallTool connects to an MCP server, calls a tool with the given params, and returns the result.
func CallTool(url, toolName string, params map[string]interface{}) (*mcp.CallToolResult, error) {
	ctx := context.Background()

	c, err := connect(ctx, url)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	req := mcp.CallToolRequest{}
	req.Params.Name = toolName
	req.Params.Arguments = params

	result, err := c.CallTool(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("tool call failed: %w", err)
	}
	return result, nil
}
