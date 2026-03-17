import asyncio
from mcp.client.streamable_http import streamable_http_client
from mcp.client.session import ClientSession
import subprocess
import time
import httpx

async def test_calculator():
    print("Testing Calculator MCP (streamable-http)...")
    
    # Start the server
    process = subprocess.Popen(
        ["uv", "run", "calculator_mcp.py", "--transport", "streamable-http"],
    )
    
    try:
        # Wait for server to start
        await asyncio.sleep(2)
        
        url = "http://127.0.0.1:8000/mcp"
        
        async with streamable_http_client(url) as streams:
            read, write = streams[0], streams[1]
            async with ClientSession(read, write) as session:
                await session.initialize()
                
                tools = await session.list_tools()
                print("Calculator tools:", [t.name for t in tools.tools])
                
                result = await session.call_tool("calculate", arguments={"a": 10, "b": 5, "operation": "add"})
                print("Calculate 10 + 5 result:", result.content)
                
                result2 = await session.call_tool("calculate", arguments={"a": 10, "b": 0, "operation": "div"})
                print("Calculate 10 / 0 result:", result2.content)
                
    finally:
        process.terminate()
        process.wait()

async def test_matrix():
    print("Testing Matrix MCP (streamable-http)...")
    
    # Start the server
    process = subprocess.Popen(
        ["uv", "run", "matrix_mcp.py", "--transport", "streamable-http"],
    )
    
    try:
        # Wait for server to start
        await asyncio.sleep(2)
        
        url = "http://127.0.0.1:8000/mcp"
        
        async with streamable_http_client(url) as streams:
            read, write = streams[0], streams[1]
            async with ClientSession(read, write) as session:
                await session.initialize()
                
                tools = await session.list_tools()
                print("Matrix tools:", [t.name for t in tools.tools])
                
                result = await session.call_tool("process_matrix", arguments={"a": [1, 2], "b": [3, 4], "operation": "add"})
                print("Matrix [1, 2] + [3, 4] result:", result.content)
                
                result2 = await session.call_tool("process_matrix", arguments={"a": [[1, 2], [3, 4]], "b": [[5, 6], [7, 8]], "operation": "dot"})
                print("Matrix dot product result:", result2.content)
                
    finally:
        process.terminate()
        process.wait()

if __name__ == "__main__":
    asyncio.run(test_calculator())
    print("-" * 40)
    asyncio.run(test_matrix())
