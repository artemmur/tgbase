CREATE TABLE tgbase_posts_raw
(
    id  Int64,
    ts  DateTime32,
    raw String CODEC (ZSTD(16))
) ENGINE = ReplacingMergeTree
      PARTITION BY toYYYYMMDD(ts)
      ORDER BY (ts, id);
