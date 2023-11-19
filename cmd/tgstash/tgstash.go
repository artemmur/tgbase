package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"tgbase/cmd/tgstash/app"
)

var (
	root           string
	clickHouseAddr string
	clickHouseDB   string
)

func init() {
	flag.StringVar(&root, "root", "", "specify buffer folder")
	flag.StringVar(&clickHouseAddr, "chAddr", "", "clickHouse addr")
	flag.StringVar(&clickHouseDB, "chDB", "", "clickHouse DB")

	flag.Parse()
	if root == "" {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	lg := slog.Default()

	svc := &app.Service{
		Logger:         lg,
		ClickHouseDB:   "",
		ClickHouseAddr: "localhost:9000",
	}

	go app.ListenForPosts(ctx, root, svc)
	go app.PushToDatabase(ctx, svc)
	<-make(chan struct{})
}
