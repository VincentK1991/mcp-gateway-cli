package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	clientName    = "gateway-cli"
	clientVersion = "1.0.0"
)

// initialize sends the MCP Initialize handshake on an already-started client.
func initialize(ctx context.Context, c *client.Client) error {
	req := mcp.InitializeRequest{}
	req.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	req.Params.ClientInfo = mcp.Implementation{Name: clientName, Version: clientVersion}
	if _, err := c.Initialize(ctx, req); err != nil {
		c.Close()
		return fmt.Errorf("failed to initialize MCP session: %w", err)
	}
	return nil
}

// connectHTTP creates and initializes a streamable HTTP client.
func connectHTTP(ctx context.Context, url string, headers map[string]string) (*client.Client, error) {
	var opts []transport.StreamableHTTPCOption
	if len(headers) > 0 {
		opts = append(opts, transport.WithHTTPHeaders(headers))
	}

	c, err := client.NewStreamableHttpClient(url, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client for %s: %w", url, err)
	}

	if err := c.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start HTTP client: %w", err)
	}

	if err := initialize(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// connectStdio launches a subprocess and initializes a stdio MCP client.
// env is a slice of "KEY=VALUE" pairs to add to the subprocess environment.
func connectStdio(ctx context.Context, command string, args []string, env []string) (*client.Client, error) {
	c, err := client.NewStdioMCPClient(command, env, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to start stdio process '%s': %w", command, err)
	}

	if err := initialize(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// FetchTools connects to an HTTP MCP server and returns its available tools.
func FetchTools(url string, headers map[string]string) (map[string]mcp.Tool, error) {
	ctx := context.Background()
	c, err := connectHTTP(ctx, url, headers)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	return listTools(ctx, c)
}

// FetchToolsStdio launches a stdio MCP subprocess and returns its available tools.
func FetchToolsStdio(command string, args []string, env []string) (map[string]mcp.Tool, error) {
	ctx := context.Background()
	c, err := connectStdio(ctx, command, args, env)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	return listTools(ctx, c)
}

// CallTool connects to an HTTP MCP server, calls a tool, and returns the result.
func CallTool(url, toolName string, params map[string]interface{}, headers map[string]string) (*mcp.CallToolResult, error) {
	ctx := context.Background()
	c, err := connectHTTP(ctx, url, headers)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	return callTool(ctx, c, toolName, params)
}

// CallToolStdio launches a stdio MCP subprocess, calls a tool, and returns the result.
func CallToolStdio(command string, args []string, env []string, toolName string, params map[string]interface{}) (*mcp.CallToolResult, error) {
	ctx := context.Background()
	c, err := connectStdio(ctx, command, args, env)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	return callTool(ctx, c, toolName, params)
}

func listTools(ctx context.Context, c *client.Client) (map[string]mcp.Tool, error) {
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

func callTool(ctx context.Context, c *client.Client, toolName string, params map[string]interface{}) (*mcp.CallToolResult, error) {
	req := mcp.CallToolRequest{}
	req.Params.Name = toolName
	req.Params.Arguments = params
	result, err := c.CallTool(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("tool call failed: %w", err)
	}
	return result, nil
}
