Yes — and for remote/streamable HTTP MCPs this is actually the **better default**. You don't want to ship a binary that has a stale snapshot of a remote MCP's schema when that MCP can change independently.

## The Problem With Embed For Remote MCPs

The embed approach assumes your gateway owns all the schemas. But if you're proxying remote streamable HTTP MCPs (Composio, hosted Notion MCP, etc.), those schemas can change without you redeploying the gateway. The binary would lie about what tools are available.

## Runtime Registration Pattern

Instead of embed, the CLI fetches schema on first use and caches it locally with a TTL:

```
First run:
  gateway-cli [anything]
    → cache miss
    → fetch /mcp/tools from gateway
    → write ~/.gateway-cli/schema-cache.json
    → register cobra commands dynamically
    → execute command

Subsequent runs (within TTL):
  gateway-cli [anything]
    → cache hit (fresh)
    → register cobra commands from cache
    → execute command

After TTL expires or --refresh:
  gateway-cli [anything]
    → cache stale
    → fetch /mcp/tools
    → update cache
    → register + execute
```

The cache TTL means startup stays fast (no network on every invocation) while the schema stays fresh enough for remote MCPs.## How It All Fits Together

The critical insight here is **when** in Cobra's lifecycle you can safely fetch the schema:

```
cobra.OnInitialize     ← too early, flags not parsed yet (no --gateway-url)
PersistentPreRunE      ← just right, flags resolved, before RunE executes  
RunE                   ← works but schema load happens inside every command
```

`PersistentPreRunE` on the root command is the right hook — it fires once, after all flags and env vars are resolved via Viper, before any subcommand runs.

## The Three Scenarios Handled

**Normal use** — cache is fresh, zero network calls, sub-100ms startup:
```bash
gateway-cli notion search-pages --query "Q3"
# reads ~/.gateway-cli/schema-cache.json, no HTTP
```

**Remote MCP schema changed** — TTL expired, transparent refresh:
```bash
gateway-cli slack post-message --channel "#x" --text "hi"
# cache stale → fetches /mcp/tools → updates cache → runs command
```

**Gateway unreachable** (network issue, deploying) — graceful degradation:
```bash
gateway-cli notion search-pages --query "Q3"
# warn: could not refresh schema (connection refused), using cached schema from 2026-03-15T10:00:00Z
# → still works using last known schema
```

**Explicit management** — for scripting and CI:
```bash
gateway-cli schema refresh          # force fetch, print new schema
gateway-cli schema info             # age, tool count, gateway URL
gateway-cli schema invalidate       # nuke cache

gateway-cli --refresh-schema notion search-pages --query "Q3"  # one-off refresh
gateway-cli --offline notion search-pages --query "Q3"         # never fetch
```

The `--offline` flag is particularly useful for your CI pipelines — you can pre-warm the cache in a setup step and then run all commands without any network dependency on the gateway being up.