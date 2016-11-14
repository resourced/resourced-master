CREATE CUSTOM INDEX IF NOT EXISTS idx_hosts_lucene ON hosts (lucene) USING 'com.stratio.cassandra.lucene.Index' WITH OPTIONS = {
    'schema' : '{
        fields : {
            id : {type : "integer"},
            cluster_id: {type : "integer"},
            hostname : {type : "string"},
            updated: {type : "integer"},
            tags : {type : "string"},
            master_tags : {type : "string"},
            data_string : { type : "string", column: "data"},
            data_float : { type : "float", column: "data"}
        }
    }'
};
