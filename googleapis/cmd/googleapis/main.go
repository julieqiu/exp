// googleapis provides tools for cataloging GitHub organizations.
//
// Usage:
//
//	go run ./cmd/googleapis catalog team --all
//	go run ./cmd/googleapis catalog team [name]
//	go run ./cmd/googleapis catalog repo --all
//	go run ./cmd/googleapis catalog repo [name]
package main

import (
	"context"
	"os"

	"github.com/julieqiu/exp/googleapis/internal/cli"
)

func main() {
	if err := cli.Run(context.Background(), os.Args); err != nil {
		os.Exit(1)
	}
}
