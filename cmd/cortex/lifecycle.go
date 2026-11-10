package main

import (
	"context"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func serve(srv *mcp.Server, dbPath string) error {
	original, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go stopOnDeletion(ctx, cancel, dbPath, original)
	err = srv.Run(ctx, &mcp.StdioTransport{})
	if ctx.Err() != nil {
		return nil
	}
	return err
}

func stopOnDeletion(ctx context.Context, cancel context.CancelFunc, path string, original os.FileInfo) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			current, err := os.Stat(path)
			if os.IsNotExist(err) || (err == nil && !os.SameFile(original, current)) {
				cancel()
				return
			}
		}
	}
}
