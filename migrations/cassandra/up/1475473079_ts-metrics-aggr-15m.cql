CREATE TABLE IF NOT EXISTS ts_metrics_aggr_15m (
    cluster_id bigint,
    metric_id bigint,
    created bigint,
    key text,
    host text,
    avg double,
    max double,
    min double,
    sum double,
    PRIMARY KEY (cluster_id, metric_id, created, key, host)
) WITH CLUSTERING ORDER BY (metric_id ASC, created DESC)
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
