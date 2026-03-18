from mcp.server.fastmcp import FastMCP
from pydantic import BaseModel, Field
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.responses import Response

mcp = FastMCP("calculator",port=8000,json_response=True,stateless_http=True)

REQUIRED_API_KEY = "test1234"

class APIKeyMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request, call_next):
        if request.headers.get("x-api-key") != REQUIRED_API_KEY:
            return Response("Unauthorized: missing or invalid x-api-key header", status_code=401)
        return await call_next(request)

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
    if args.transport == "streamable-http":
        import uvicorn
        app = mcp.streamable_http_app()
        app.add_middleware(APIKeyMiddleware)
        uvicorn.run(app, host=mcp.settings.host, port=mcp.settings.port, log_level=mcp.settings.log_level.lower())
    else:
        mcp.run(transport=args.transport) # type: ignore
