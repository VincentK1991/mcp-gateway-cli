from mcp.server.fastmcp import FastMCP
from pydantic import BaseModel, Field
import numpy as np
from typing import Any

mcp = FastMCP("matrix",port=8001,json_response=True,stateless_http=True)

class MatrixResult(BaseModel):
    result: list[Any] | float | None = Field(default=None, description="The matrix or scalar result")
    error: str | None = Field(default=None, description="Error message if any")

@mcp.tool()
def process_matrix(a: list[Any] | float, b: list[Any] | float, operation: str) -> MatrixResult:
    """Perform matrix operations on 1D/2D arrays or scalars.
    
    Args:
        a: First matrix/scalar
        b: Second matrix/scalar
        operation: 'add', 'sub', 'mul' (element-wise), or 'dot' (matrix multiplication)
    """
    try:
        arr_a = np.array(a)
        arr_b = np.array(b)

        if operation == "add":
            res = arr_a + arr_b
        elif operation == "sub":
            res = arr_a - arr_b
        elif operation == "mul":
            res = arr_a * arr_b
        elif operation == "dot":
            res = np.dot(arr_a, arr_b)
        else:
            return MatrixResult(error=f"Unknown operation: {operation}")
            
        # Convert numpy array/scalar back to basic python types
        if isinstance(res, (int, float, np.number)):
            return MatrixResult(result=float(res)) # type: ignore
        else:
            return MatrixResult(result=res.tolist())  # type: ignore[attr-defined]

    except Exception as e:
        return MatrixResult(error=str(e))

if __name__ == "__main__":
    import argparse
    parser = argparse.ArgumentParser(description="Run Matrix MCP Server")
    parser.add_argument("--transport", default="stdio", choices=["stdio", "streamable-http"], help="Transport protocol to use")
    args = parser.parse_args()
    mcp.run(transport=args.transport) # type: ignore
