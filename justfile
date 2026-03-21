mcps_dir := "mcps"
pids_file := ".mcps.pids"

# ── Go ──────────────────────────────────────────────────────────────────────

# Build the gateway-cli binary
build:
    go build -o gateway-cli .

# Run all Go tests
test:
    go test -v ./...

# Run go vet
lint:
    go vet ./...

# ── MCP servers (streamable-http) ───────────────────────────────────────────

# Start all MCP servers in the background (streamable-http)
mcps-start:
    #!/usr/bin/env bash
    set -e
    rm -f {{pids_file}}
    for script in calculator_mcp.py matrix_mcp.py large_output_mcp.py; do
        echo "Starting $script …"
        (cd {{mcps_dir}} && uv run python "$script" --transport streamable-http \
            > "/tmp/${script%.py}.log" 2>&1) &
        echo $! >> {{pids_file}}
    done
    echo "All MCP servers started. PIDs in {{pids_file}}"

# Stop all MCP servers started by mcps-start
mcps-stop:
    #!/usr/bin/env bash
    if [ ! -f {{pids_file}} ]; then
        echo "No PID file found ({{pids_file}})"
        exit 0
    fi
    while read -r pid; do
        kill "$pid" 2>/dev/null && echo "Stopped PID $pid" || true
    done < {{pids_file}}
    rm -f {{pids_file}}

# Show running status of MCP servers
mcps-status:
    #!/usr/bin/env bash
    if [ ! -f {{pids_file}} ]; then
        echo "No PID file found."
        exit 0
    fi
    while read -r pid; do
        if kill -0 "$pid" 2>/dev/null; then
            echo "PID $pid  running"
        else
            echo "PID $pid  stopped"
        fi
    done < {{pids_file}}

# Tail logs from all MCP servers
mcps-logs:
    #!/usr/bin/env bash
    for script in calculator_mcp.py matrix_mcp.py large_output_mcp.py; do
        log="/tmp/${script%.py}.log"
        echo "=== $script ==="
        [ -f "$log" ] && cat "$log" || echo "(no log yet)"
    done
