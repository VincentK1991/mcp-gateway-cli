"""MCP server for testing large-output handling in the gateway-cli.

The gateway-cli writes output to a temp file when it exceeds 40,000 chars.
Tools here produce output above and below that threshold.
"""
import random
import string

from mcp.server.fastmcp import FastMCP

mcp = FastMCP("large-output-test", port=8002, json_response=True, stateless_http=True)

THRESHOLD = 40_000


def _random_word(length: int = 8) -> str:
    return "".join(random.choices(string.ascii_lowercase, k=length))


@mcp.tool()
def large_text(chars: int = 50_000) -> str:
    """Return a plain-text string of approximately `chars` characters.

    Default (50 000) exceeds the 40 000-char gateway-cli threshold so the
    output is written to a temp file instead of flooding stdout.

    Args:
        chars: Approximate number of characters to return. Use a value > 40000
               to trigger file-writing; use a value <= 40000 to stay below it.
    """
    words = []
    total = 0
    while total < chars:
        w = _random_word(random.randint(4, 12))
        words.append(w)
        total += len(w) + 1  # +1 for space/newline
    return " ".join(words)


@mcp.tool()
def large_json(num_records: int = 500) -> list[dict]:
    """Return a JSON array of `num_records` objects.

    Each record has an id, name, value, and tags list.
    500 records ≈ 60 000+ chars — above the 40 000-char threshold.
    Use ~300 records to stay below threshold.

    Args:
        num_records: Number of records to generate.
    """
    return [
        {
            "id": i,
            "name": _random_word(10),
            "value": round(random.uniform(0, 10_000), 4),
            "tags": [_random_word(6) for _ in range(random.randint(2, 6))],
            "description": " ".join(_random_word(8) for _ in range(10)),
        }
        for i in range(num_records)
    ]


@mcp.tool()
def nested_json(depth: int = 5, breadth: int = 4) -> dict:
    """Return a deeply nested JSON object to test large structured output.

    Args:
        depth: How many levels deep to nest.
        breadth: Number of keys at each level.
    """

    def build(d: int) -> dict:
        if d == 0:
            return {_random_word(5): _random_word(20) for _ in range(breadth)}
        node: dict = {_random_word(5): _random_word(15) for _ in range(breadth // 2)}
        node["children"] = [build(d - 1) for _ in range(breadth)]
        return node

    return build(depth)


@mcp.tool()
def small_text(chars: int = 100) -> str:
    """Return a short text string well below the 40 000-char threshold.

    Useful as a baseline / sanity-check tool.

    Args:
        chars: Number of characters to return (should be < 40000).
    """
    words = []
    total = 0
    while total < chars:
        w = _random_word(random.randint(4, 8))
        words.append(w)
        total += len(w) + 1
    return " ".join(words)


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="Run Large-Output Test MCP Server")
    parser.add_argument(
        "--transport",
        default="stdio",
        choices=["stdio", "streamable-http"],
        help="Transport protocol to use",
    )
    args = parser.parse_args()
    if args.transport == "streamable-http":
        import uvicorn

        app = mcp.streamable_http_app()
        uvicorn.run(
            app,
            host=mcp.settings.host,
            port=mcp.settings.port,
            log_level=mcp.settings.log_level.lower(),
        )
    else:
        mcp.run(transport=args.transport)  # type: ignore
