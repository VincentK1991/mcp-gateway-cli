from mcp.server.fastmcp import FastMCP
from pydantic import BaseModel, Field

mcp = FastMCP("calculator",port=8000,json_response=True,stateless_http=True)

class OperationResult(BaseModel):
    result: float | None = Field(default=None, description="The result of the operation")
    error: str | None = Field(default=None, description="Error message if any")

@mcp.tool()
def calculate(a: int, b: int, operation: str) -> OperationResult:
    """Perform a basic integer calculation.
    
    Args:
        a: First integer operand
        b: Second integer operand
        operation: 'add', 'sub', 'mul', or 'div'
    """
    if operation == "add":
        return OperationResult(result=float(a + b))
    elif operation == "sub":
        return OperationResult(result=float(a - b))
    elif operation == "mul":
        return OperationResult(result=float(a * b))
    elif operation == "div":
        if b == 0:
            return OperationResult(error="Division by zero")
        return OperationResult(result=float(a) / float(b))
    else:
        return OperationResult(error=f"Unknown operation: {operation}")

if __name__ == "__main__":
    import argparse
    parser = argparse.ArgumentParser(description="Run Calculator MCP Server")
    parser.add_argument("--transport", default="stdio", choices=["stdio", "streamable-http"], help="Transport protocol to use")
    args = parser.parse_args()
    mcp.run(transport=args.transport) # type: ignore
