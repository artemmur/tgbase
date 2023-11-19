package tgbase

import "time"

type Post struct {
	ID        int64
	CreatedAt time.Time
	Message   string
	ChannelID int64
}
