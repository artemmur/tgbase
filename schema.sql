CREATE TABLE tgbase_posts_raw
(
    chatID Int64,
    msgID  Int64,
    ts  DateTime32,
    raw String CODEC (ZSTD(16))
) ENGINE = ReplacingMergeTree
      PARTITION BY toYYYYMMDD(ts)
      ORDER BY (ts, chatID, msgID);
