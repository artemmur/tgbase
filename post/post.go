package post

import (
	"encoding/json"
	"os"
)

type Post struct {
	Date      int
	Message   string
	ChannelID int64
}

func FlushPost(p *Post, destFolder string) error {
	f, err := os.CreateTemp(destFolder, "*.json")
	if err != nil {
		return err
	}

	if err := json.NewEncoder(f).Encode(&p); err != nil {
		return err
	}
	return nil
}
