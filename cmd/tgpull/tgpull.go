package main

import (
	"context"
	"os"
	"os/signal"

	"tgbase/cmd/tgpull/app"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := app.StartObserver(ctx); err != nil {
		panic(err)
	}
}
