package schema

import "time"

// GatewaySchema is the top-level structure holding all cached MCP tool schemas.
type GatewaySchema struct {
	LastFetch time.Time      `json:"lastFetch"`
	MCPs      map[string]MCP `json:"mcps"`
}

// MCP represents a single MCP server and its tools.
type MCP struct {
	Name  string          `json:"name"`
	Tools map[string]Tool `json:"tools"`
}

// Tool represents a callable tool exposed by an MCP server.
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema describes the parameters a Tool accepts.
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

// Property describes a single input parameter.
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

func newGatewaySchema() *GatewaySchema {
	return &GatewaySchema{MCPs: make(map[string]MCP)}
}
