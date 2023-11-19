package app

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"tgbase"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/updates"
	updhook "github.com/gotd/td/telegram/updates/hook"
	"github.com/gotd/td/tg"
)

func flushPost(destFolder string, timeout time.Duration, size int) func(p *tgbase.Post) {
	collection := make([]tgbase.Post, 0, size)
	save := func() error {
		defer func() {
			collection = make([]tgbase.Post, 0, size)
		}()

		f, err := os.CreateTemp(destFolder, "*.json")
		if err != nil {
			return err
		}

		if err := json.NewEncoder(f).Encode(&collection); err != nil {
			return err
		}
		return nil
	}

	tick := time.NewTicker(timeout)
	pipe := make(chan tgbase.Post)
	go func() {
		for {
			select {
			case <-tick.C:
				if len(collection) > 0 {
					if err := save(); err != nil {
						slog.Error(err.Error())
					}
				}

			case newPost := <-pipe:
				collection = append(collection, newPost)
				if len(collection) == size {
					save()
				}
			}
		}
	}()

	return func(p *tgbase.Post) {
		tick.Reset(timeout)
		pipe <- *p
	}
}

func StartObserver(ctx context.Context, root string) error {
	d := tg.NewUpdateDispatcher()
	gaps := updates.New(updates.Config{
		Handler: d,
	})

	// Authentication flow handles authentication process, like prompting for code and 2FA password.
	phone := os.Getenv("TG_PHONE")
	codePrompt := func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
		// NB: Use "golang.org/x/crypto/ssh/terminal" to prompt password.
		fmt.Print("Enter code: ")
		code, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(code), nil
	}
	flow := auth.NewFlow(
		auth.Constant(phone, "", auth.CodeAuthenticatorFunc(codePrompt)),
		auth.SendCodeOptions{},
	)

	// Initializing client from environment.
	// Available environment variables:
	// 	APP_ID:         app_id of Telegram app.
	// 	APP_HASH:       app_hash of Telegram app.
	// 	SESSION_FILE:   path to session file
	// 	SESSION_DIR:    path to session directory, if SESSION_FILE is not set
	client, err := telegram.ClientFromEnvironment(telegram.Options{
		UpdateHandler: gaps,
		Middlewares: []telegram.Middleware{
			updhook.UpdateHook(gaps.Handle),
		},
	})
	if err != nil {
		return err
	}

	fp := flushPost(root, 3*time.Second, 400)
	// Setup message update handlers.
	d.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewChannelMessage) error {
		msg, ok := update.Message.(*tg.Message)
		if !ok {
			return nil
		}

		p := &tgbase.Post{
			MessageID: int64(msg.ID),
			CreatedAt: time.Unix(int64(msg.Date), 0),
			Message:   msg.Message,
		}

		peer, ok := msg.PeerID.(*tg.PeerChannel)
		if ok {
			p.ChannelID = peer.ChannelID
		}

		fp(p)
		return nil
	})

	return client.Run(ctx, func(ctx context.Context) error {
		// Perform auth if no session is available.
		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			return fmt.Errorf("%w: %v", err, "auth")
		}

		// Fetch user info.
		user, err := client.Self(ctx)
		if err != nil {
			return fmt.Errorf("%w: %v", err, "call self")
		}

		return gaps.Run(ctx, client.API(), user.ID, updates.AuthOptions{
			OnStart: func(ctx context.Context) {
				slog.Info("Gaps started")
			},
		})
	})
}
