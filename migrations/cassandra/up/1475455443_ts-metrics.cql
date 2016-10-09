CREATE TABLE IF NOT EXISTS ts_metrics (
    cluster_id bigint,
    metric_id bigint,
    created bigint,
    key text,
    host text,
    value double,
    PRIMARY KEY ((cluster_id, metric_id), created)
) WITH CLUSTERING ORDER BY (created ASC)
  AND compaction = {'compaction_window_unit': 'HOURS', 'compaction_window_size': '24', 'class':'com.jeffjirsa.cassandra.db.compaction.TimeWindowCompactionStrategy'}
  AND caching = {'keys': 'NONE', 'rows_per_partition': 'NONE'}
  AND compression = {'sstable_compression': ''}
  AND crc_check_chance = 0
  AND dclocal_read_repair_chance = 0
  AND default_time_to_live = 0
  AND gc_grace_seconds = 0
  AND memtable_flush_period_in_ms = 0
  AND read_repair_chance = 0.0
  AND speculative_retry = '99PERCENTILE';
