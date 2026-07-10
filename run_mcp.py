"""
Atalho para subir o servidor MCP (stdio).
    python run_mcp.py
"""

from mtga.mcp_server import mcp

if __name__ == "__main__":
    mcp.run()
