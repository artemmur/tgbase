package post

import (
	"encoding/json"
	"os"
)

type Post struct {
	Date    int
	Message string
}

func FlushPost(p *Post, destFolder string) error {
	f, err := os.CreateTemp(destFolder, "*.txt")
	if err != nil {
		return err
	}

	if err := json.NewEncoder(f).Encode(&p); err != nil {
		return err
	}
	return nil
}
