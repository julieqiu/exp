package main

import (
	"context"
	"fmt"
	"os"

	"github.com/julieqiu/exp/surfer/internal/surfer"
)

func main() {
	if err := surfer.Run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
