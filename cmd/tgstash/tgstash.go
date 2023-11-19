package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"tgbase/cmd/tgstash/app"

	"github.com/ClickHouse/ch-go"
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

	svc := app.NewService(lg, &app.ConfigSvc{
		ClickHouseDB:   clickHouseDB,
		ClickHouseAddr: clickHouseAddr,
	})

	conn, err := ch.Dial(ctx, ch.Options{
		Database: clickHouseDB,
		Address:  clickHouseAddr,
		User:     os.Getenv("CLICKHOUSE_USER"),
		Password: os.Getenv("CLICKHOUSE_PASSWORD"),
	})
	if err != nil {
		panic(err)
	}
	if err := conn.Ping(ctx); err != nil {
		panic(err)
	}
	conn.Close()

	go app.ListenForPosts(ctx, root, svc)
	go app.PushToDatabase(ctx, svc)

	<-ctx.Done()
}
