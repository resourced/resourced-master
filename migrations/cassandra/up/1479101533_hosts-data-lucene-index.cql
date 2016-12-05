CREATE CUSTOM INDEX IF NOT EXISTS idx_hosts_data_lucene ON hosts_data (lucene) USING 'com.stratio.cassandra.lucene.Index' WITH OPTIONS = {
    'schema' : '{
        fields : {
            id : { type : "string"},
            key : { type : "string"},
            value_string : { type : "string", column: "value"},
            value_float : { type : "float", column: "value"}
        }
    }'
};
