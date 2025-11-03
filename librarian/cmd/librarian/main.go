package main

import (
	"context"
	"log"
	"os"

	"github.com/julieqiu/exp/librarian/internal/librarian"
)

func main() {
	if err := librarian.NewApp().Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}