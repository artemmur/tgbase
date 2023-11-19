package tgbase

import "time"

type Post struct {
	ChannelID int64
	MessageID int64
	CreatedAt time.Time
	Message   string
}
