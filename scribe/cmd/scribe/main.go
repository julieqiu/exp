package main

import (
	"context"
	"log"
	"os"

	"github.com/julieqiu/exp/scribe/internal/scribe"
)

func main() {
	if err := scribe.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
