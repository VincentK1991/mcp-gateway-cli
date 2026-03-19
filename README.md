# gateway-cli

A command-line tool for interacting with [MCP](https://modelcontextprotocol.io) servers. It discovers tools exposed by registered MCP servers, caches their schemas, and lets you call them directly from your terminal — with full support for piping, scripting, and composition with other CLI tools.

---

## Installation

### Homebrew (macOS and Linux)

```bash
brew install VincentK1991/tap/gateway-cli
```

### curl (macOS and Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/VincentK1991/mcp-gateway-cli/main/install.sh | bash
```

This detects your OS and architecture, downloads the correct binary from the latest GitHub release, verifies the checksum, and installs to `/usr/local/bin`.

To install a specific version:

```bash
VERSION=v1.2.0 curl -fsSL https://raw.githubusercontent.com/VincentK1991/mcp-gateway-cli/main/install.sh | bash
```

### Go install (requires Go 1.23+)

```bash
go install github.com/VincentK1991/mcp-gateway-cli@latest
```

### Verify installation

```bash
gateway-cli --version
```

---

## Registering MCP Servers

MCP servers are registered in `~/.gateway-cli/config.yaml`. Create the file if it doesn't exist:

```yaml
mcps:
  calculator:
    url: http://localhost:8000/mcp
  matrix:
    url: http://localhost:8001/mcp
```

Each entry needs a name (used as the CLI subcommand) and the URL of the MCP server's `/mcp` endpoint.

For MCP servers that require authentication or custom headers, add a `headers` block:

```yaml
mcps:
  calculator:
    url: http://localhost:8000/mcp
  my-api:
    url: https://api.example.com/mcp
    headers:
      Authorization: "Bearer mytoken"
  another-api:
    url: https://other.example.com/mcp
    headers:
      Authorization: "Bearer ${MY_API_TOKEN}"   # reads from environment variable
      X-Tenant-Id: "acme"
```

Header values support environment variable expansion — use `${VAR}` to avoid storing secrets directly in the config file. Set the variable in your shell before running any command:

```bash
export MY_API_TOKEN="your-secret-token"
gateway-cli my-api some-tool --flag value
```

After editing the config, refresh the schema cache:

```bash
gateway-cli schema refresh
```

You can inspect what was discovered:

```bash
gateway-cli schema info
```

```
Last fetched : 2026-03-16 22:29:12
MCP servers  : 2
Total tools  : 2

  calculator (1 tools)
    calculate            Perform a basic integer calculation.

  matrix (1 tools)
    process_matrix       Perform matrix operations on 1D/2D arrays or scalars.
```

---

## Running Tools

Tools are invoked as: `gateway-cli <mcp-name> <tool-name> [flags]`

Each tool's flags are generated automatically from its input schema. Use `--help` to see available flags:

```bash
gateway-cli calculator calculate --help
gateway-cli matrix process_matrix --help
```

**Basic examples:**

```bash
# Integer arithmetic
gateway-cli calculator calculate --a 10 --b 5 --operation add
gateway-cli calculator calculate --a 10 --b 3 --operation mul
gateway-cli calculator calculate --a 10 --b 0 --operation div   # returns error gracefully

# Vector operations
gateway-cli matrix process_matrix --a '[1,2,3]' --b '[4,5,6]' --operation add

# Matrix dot product
gateway-cli matrix process_matrix --a '[[1,2],[3,4]]' --b '[[5,6],[7,8]]' --operation dot
```

---

## Composing Tool Outputs

Use the `--json` / `-j` flag to strip the MCP envelope and return only the structured result. This makes chaining tools together straightforward:

```bash
# Step 1: compute a scale factor
# Step 2: use it as input to the matrix tool
SCALE=$(gateway-cli --json calculator calculate --a 3 --b 4 --operation mul | jq '.result | floor')
gateway-cli --json matrix process_matrix --a '[[1,2],[3,4]]' --b "$SCALE" --operation mul
```

```json
{
  "error": null,
  "result": [[12, 24], [36, 48]]
}
```

Multi-step pipelines:

```bash
# Chain three calculator calls
A=$(gateway-cli --json calculator calculate --a 2 --b 3 --operation add | jq '.result | floor')
B=$(gateway-cli --json calculator calculate --a 4 --b 2 --operation mul | jq '.result | floor')
gateway-cli --json calculator calculate --a "$A" --b "$B" --operation add | jq '.result'
# → 13
```

---

## Combining with Other CLI Tools

The output is standard JSON, so it integrates naturally with the Unix toolchain.

**jq — extract and transform:**

```bash
# Extract just the result value
gateway-cli --json calculator calculate --a 10 --b 3 --operation mul | jq '.result'

# Extract matrix rows
gateway-cli --json matrix process_matrix --a '[[1,2],[3,4]]' --b '[[5,6],[7,8]]' --operation dot \
  | jq '.result[]'
```

**Save output to a file:**

```bash
gateway-cli --json matrix process_matrix --a '[[1,2],[3,4]]' --b '[[5,6],[7,8]]' --operation dot \
  > result.json
```

**Use in shell scripts:**

```bash
#!/bin/bash
for op in add sub mul; do
  echo -n "$op: "
  gateway-cli --json calculator calculate --a 10 --b 3 --operation "$op" | jq '.result'
done
```

**Pipe into other tools (`fx`, `python`, etc.):**

```bash
# Interactive JSON explorer
gateway-cli --json matrix process_matrix --a '[[1,2],[3,4]]' --b '[[5,6],[7,8]]' --operation dot | fx

# Process in Python
gateway-cli --json calculator calculate --a 10 --b 3 --operation add \
  | python3 -c "import sys, json; d = json.load(sys.stdin); print(f'Result is {d[\"result\"]}')"
```

---

## Schema Cache Management

```bash
gateway-cli schema refresh      # force re-fetch from all MCP servers
gateway-cli schema info         # show cached MCPs and tools
gateway-cli schema invalidate   # delete the cache
```

Flags available on every command:

| Flag | Description |
|---|---|
| `--json`, `-j` | Output structured content only (no MCP envelope) |
| `--text`, `-t` | Output the first text content item as a plain string |
| `--refresh-schema` | Re-fetch schema before running the command |
| `--offline` | Use cached schema only, never contact MCP servers |
| `--config` | Path to a custom config file |

### Checking and updating your version

```bash
# Show installed version
gateway-cli --version

# Homebrew
brew upgrade gateway-cli

# curl (re-runs the installer, overwrites the binary)
curl -fsSL https://raw.githubusercontent.com/VincentK1991/mcp-gateway-cli/main/install.sh | bash

# Go install
go install github.com/VincentK1991/mcp-gateway-cli@latest
```

gateway-cli also checks for new releases automatically in the background (at most once per day) and prints a reminder to stderr when one is available.
