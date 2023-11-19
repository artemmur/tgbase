package main

import (
	"context"
	"flag"
	"os"
	"os/signal"

	"tgbase/cmd/tgpull/app"
)

var root string

func init() {
	flag.StringVar(&root, "root", "", "root folder to flush post from channels")
	flag.Parse()
	if root == "" {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := app.StartObserver(ctx, root); err != nil {
		panic(err)
	}
}
