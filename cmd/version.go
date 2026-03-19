package cmd

// Version is set at build time via:
//
//	go build -ldflags "-X github.com/VincentK1991/mcp-gateway-cli/cmd.Version=v1.0.0"
//
// When built without ldflags (e.g. during development), it defaults to "dev".
var Version = "dev"
