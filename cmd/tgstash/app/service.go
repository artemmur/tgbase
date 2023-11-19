package app

import (
	"log/slog"
	"tgbase"
)

type Service struct {
	batches chan []tgbase.Post

	ClickHouseDB   string
	ClickHouseAddr string
	Logger         *slog.Logger
}
