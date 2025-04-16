package mcp

import (
	"context"

	"github.com/daodao97/xgo/xapp"
	"github.com/daodao97/xgo/xlog"
	"github.com/mark3labs/mcp-go/server"
)

type MCP struct {
	addr      string
	sseServer *server.SSEServer
}

func NewMcpServer(addr string, sseServer *server.SSEServer) xapp.NewServer {
	return func() xapp.Server {
		return &MCP{
			addr:      addr,
			sseServer: sseServer,
		}
	}
}

func (m *MCP) Start() error {
	xlog.Debug("MCP server started on", xlog.String("addr", m.addr))
	return m.sseServer.Start(m.addr)
}

func (m *MCP) Stop() {
	xlog.Debug("MCP server stopped")
	m.sseServer.Shutdown(context.Background())
}
