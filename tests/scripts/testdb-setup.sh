#!/bin/bash

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
MIGRATIONS_DIR=$CURRENT_DIR/../../migrations

echo "Setting up PostgreSQL and Cassandra databases for tests. You must have both of them up and running."

for file in $MIGRATIONS_DIR/pg/up/*
do
    if [[ -f $file ]]; then
        echo "Running PostgreSQL migration: $file"
        psql -d "resourced-master-test" -f $file
    fi
done

for file in $MIGRATIONS_DIR/cassandra/up/*
do
    if [[ -f $file ]]; then
        echo "Running Cassandra migration: $file"
        cqlsh -k resourced_master_test -f $file
    fi
done
