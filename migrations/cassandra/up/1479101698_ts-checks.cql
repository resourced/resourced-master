CREATE TABLE IF NOT EXISTS ts_checks (
    cluster_id bigint,
    check_id bigint,
    created bigint,
    result boolean,
    expressions text,
    PRIMARY KEY (cluster_id, check_id, created)
) WITH CLUSTERING ORDER BY (check_id ASC, created ASC)
  AND compaction = {'compaction_window_unit': 'HOURS', 'compaction_window_size': '24', 'class':'org.apache.cassandra.db.compaction.TimeWindowCompactionStrategy'}
  AND caching = {'keys': 'NONE', 'rows_per_partition': 'NONE'}
  AND compression = {'sstable_compression' : 'LZ4Compressor'}
  AND crc_check_chance = 0
  AND dclocal_read_repair_chance = 0
  AND default_time_to_live = 0
  AND gc_grace_seconds = 0
  AND memtable_flush_period_in_ms = 0
  AND read_repair_chance = 0.0
  AND speculative_retry = '99PERCENTILE';
