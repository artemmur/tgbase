package app

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"tgbase"

	"github.com/fsnotify/fsnotify"
)

func readPostsFromFile(filename string) ([]tgbase.Post, error) {
	data, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer data.Close()

	posts := make([]tgbase.Post, 0)
	if err := json.NewDecoder(data).Decode(&posts); err != nil {
		return nil, err
	}
	return posts, nil
}

func ListenForPosts(ctx context.Context, root string, svc *Service) {
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	process := func(filename string) {
		posts, err := readPostsFromFile(filename)
		if err != nil {
			svc.logger.ErrorContext(ctx, err.Error())
			return
		}
		svc.batches <- posts

		if err := os.Remove(filename); err != nil {
			svc.logger.ErrorContext(ctx, err.Error())
			return
		}
	}

	// Start listening for events.
	go func() {
	OUTER:
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					process(filepath.Join(root, event.Name))
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			default:
				dir, err := os.ReadDir(root)
				if err != nil {
					svc.logger.ErrorContext(ctx, err.Error())
					break
				}
				if len(dir) == 0 {
					break OUTER
				}
				process(filepath.Join(root, dir[0].Name()))
			}
		}
	}()

	// Add a path.
	err = watcher.Add(root)
	if err != nil {
		log.Fatal(err)
	}

	// Block main goroutine forever.
	<-make(chan struct{})
}
