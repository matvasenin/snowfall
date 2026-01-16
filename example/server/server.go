// Copyright 2025 The Go MCP SDK Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	host  = "localhost"
	port  = 9000
	proto = "http"
)

type HelloParams struct {
	Name string `json:"name"`
}
type AddParams struct {
	A uint16 `json:"a"`
	B uint16 `json:"b"`
}

func Hello(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args HelloParams,
) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Hello, " + args.Name},
		},
	}, nil, nil
}
func Add(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args AddParams,
) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("%d", args.A+args.B)},
		},
	}, nil, nil
}

func main() {
	url := fmt.Sprintf("%s:%d", host, port)
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "example-server",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "hello",
		Description: "Says Hello, as it always was",
	}, Hello)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "add",
		Description: "Adds two numbers, as it forever will be",
	}, Add)

	handler := mcp.NewStreamableHTTPHandler(
		func(req *http.Request) *mcp.Server {
			return server
		}, nil,
	)
	log.Printf("MCP server listening on %s", url)

	if err := http.ListenAndServe(url, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
