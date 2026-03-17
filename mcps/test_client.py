import asyncio
from mcp.client.stdio import stdio_client, StdioServerParameters
from mcp.client.session import ClientSession
import os

async def test_calculator():
    print("Testing Calculator MCP (stdio)...")
    server_params = StdioServerParameters(
        command="uv",
        args=["run", "calculator_mcp.py", "--transport", "stdio"],
        env={**os.environ}
    )
    
    async with stdio_client(server_params) as (read, write):
        async with ClientSession(read, write) as session:
            await session.initialize()
            
            tools = await session.list_tools()
            print("Calculator tools:", [t.name for t in tools.tools])
            
            result = await session.call_tool("calculate", arguments={"a": 10, "b": 5, "operation": "add"})
            print("Calculate 10 + 5 result:", result.content)
            
            result2 = await session.call_tool("calculate", arguments={"a": 10, "b": 0, "operation": "div"})
            print("Calculate 10 / 0 result:", result2.content)

async def test_matrix():
    print("Testing Matrix MCP (stdio)...")
    server_params = StdioServerParameters(
        command="uv",
        args=["run", "matrix_mcp.py", "--transport", "stdio"],
        env={**os.environ}
    )
    
    async with stdio_client(server_params) as (read, write):
        async with ClientSession(read, write) as session:
            await session.initialize()
            
            tools = await session.list_tools()
            print("Matrix tools:", [t.name for t in tools.tools])
            
            result = await session.call_tool("process_matrix", arguments={"a": [1, 2], "b": [3, 4], "operation": "add"})
            print("Matrix [1, 2] + [3, 4] result:", result.content)
            
            result2 = await session.call_tool("process_matrix", arguments={"a": [[1, 2], [3, 4]], "b": [[5, 6], [7, 8]], "operation": "dot"})
            print("Matrix dot product result:", result2.content)

if __name__ == "__main__":
    asyncio.run(test_calculator())
    print("-" * 40)
    asyncio.run(test_matrix())
