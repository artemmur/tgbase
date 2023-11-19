package app

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/ClickHouse/ch-go"
	"github.com/ClickHouse/ch-go/proto"
)

func PushToDatabase(ctx context.Context, svc *Service) {
	for {
		if err := ctx.Err(); err != nil {
			svc.logger.ErrorContext(ctx, err.Error())
			return
		}
		db, err := ch.Dial(ctx, ch.Options{
			Database: svc.cfg.ClickHouseDB,
			Address:  svc.cfg.ClickHouseAddr,
			User:     os.Getenv("CLICKHOUSE_USER"),
			Password: os.Getenv("CLICKHOUSE_PASSWORD"),
		})
		if err != nil {
			svc.logger.ErrorContext(ctx, err.Error())
			return
		}

		// Restart stream every softTimeout to force merges.
		softTimeout := time.Now().Add(time.Minute)

		var (
			colChID  proto.ColInt64    // chatID Int64
			colMsgID proto.ColInt64    // msgID Int64
			colTs    proto.ColDateTime // ts DateTime
			colRaw   proto.ColBytes    // raw String
		)
		q := ch.Query{
			Body: "INSERT INTO tgbase_posts_raw VALUES",
			Input: proto.Input{
				{Name: "chatID", Data: &colChID},
				{Name: "msgID", Data: &colMsgID},
				{Name: "ts", Data: &colTs},
				{Name: "raw", Data: &colRaw},
			},
			OnInput: func(ctx context.Context) error {
				// Stream events to ClickHouse.
				colChID.Reset()
				colMsgID.Reset()
				colTs.Reset()
				colRaw.Reset()
				if time.Now().After(softTimeout) {
					// Restarting stream to force merges.
					return io.EOF
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Second * 5):
					// No events for 5 seconds, restarting stream.
					return io.EOF
				case batch := <-svc.batches:
					for _, e := range batch {
						colMsgID.Append(e.MessageID)
						colChID.Append(e.ChannelID)
						colTs.Append(e.CreatedAt)

						raw, err := json.Marshal(e)
						if err != nil {
							svc.logger.ErrorContext(ctx, err.Error())
							continue
						}
						colRaw.Append(raw)
					}
					return nil
				}
			},
		}
		if err := db.Do(ctx, q); err != nil {
			svc.logger.ErrorContext(ctx, err.Error())
			return
		}
	}
}
