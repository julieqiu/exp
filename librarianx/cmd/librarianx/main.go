package main

import (
	"context"
	"fmt"
	"os"

	"github.com/julieqiu/xlibrarian/internal/librarian"
)

func main() {
	ctx := context.Background()
	if err := librarian.Run(ctx, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "librarian: %v\n", err)
		os.Exit(1)
	}
}
