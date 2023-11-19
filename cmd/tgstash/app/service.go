package app

import (
	"log/slog"
	"tgbase"
)

type Service struct {
	batches chan []tgbase.Post
	logger  *slog.Logger
	cfg     *ConfigSvc
}

type ConfigSvc struct {
	ClickHouseDB   string
	ClickHouseAddr string
}

func NewService(lg *slog.Logger, cfg *ConfigSvc) *Service {
	return &Service{
		cfg:     cfg,
		batches: make(chan []tgbase.Post, 5),
		logger:  lg,
	}
}
