package app

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"tgbase"

	"github.com/fsnotify/fsnotify"
)

func readPostsFromFile(filename string) ([]tgbase.Post, error) {
	data, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
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

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Has(fsnotify.Write) {
					posts, err := readPostsFromFile(event.Name)
					if err != nil {
						svc.Logger.ErrorContext(ctx, err.Error())
						break
					}
					svc.batches <- posts
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
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
