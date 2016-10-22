CREATE CUSTOM INDEX idx_ts_logs_lucene ON ts_logs (lucene) USING 'com.stratio.cassandra.lucene.Index' WITH OPTIONS = {
    'schema' : '{
        fields : {
            id : {type : "integer"},
            cluster_id: {type : "integer"},
            created: {type : "integer"},
            hostname : {type : "string"},
            tags : {type : "string"},
            filename : {type : "string"},
            logline : {type : "text", analyzer : "english"}
        }
    }'
};
