package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	endpoint = flag.String("http", "", "if set, connect to this streamable endpoint rather than running a stdio server")
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 && *endpoint == "" {
		fmt.Fprintln(os.Stderr, "Usage: listfeatures <command> [<args>]")
		fmt.Fprintln(os.Stderr, "Usage: listfeatures --http=\"https://example.com/server/mcp\"")
		fmt.Fprintln(os.Stderr, "List all features for a stdio MCP server")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Example:\n\tlistfeatures go run github.com/modelcontextprotocol/go-sdk/examples/server/hello")
		os.Exit(2)
	}

	var (
		ctx       = context.Background()
		transport mcp.Transport
	)
	if *endpoint != "" {
		transport = &mcp.StreamableClientTransport{
			Endpoint: *endpoint,
		}
	} else {
		cmd := exec.Command(args[0], args[1:]...)
		transport = &mcp.CommandTransport{Command: cmd}
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "mcp-client", Version: "v1.0.0"}, nil)
	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer cs.Close()
}
