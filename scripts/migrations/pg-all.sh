#!/bin/bash
# This script is useful for development purpose only.

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

PGUSER=${PGUSER:-$(whoami)}
PGHOST=${PGHOST:-"localhost"}
PGPORT=${PGPORT:-"5432"}
PGSSLMODE=${PGSSLMODE:-"disable"}

migrate -url postgres://$PGUSER@$PGHOST:$PGPORT/resourced-master?sslmode=$PGSSLMODE -path $CURRENT_DIR/../../migrations/pg/core $@
migrate -url postgres://$PGUSER@$PGHOST:$PGPORT/resourced-master-test?sslmode=$PGSSLMODE -path $CURRENT_DIR/../../migrations/pg/core $@

migrate -url postgres://$PGUSER@$PGHOST:$PGPORT/resourced-master-hosts?sslmode=$PGSSLMODE -path $CURRENT_DIR/../../migrations/pg/hosts $@

migrate -url postgres://$PGUSER@$PGHOST:$PGPORT/resourced-master-ts-events?sslmode=$PGSSLMODE -path $CURRENT_DIR/../../migrations/pg/ts-events $@

migrate -url postgres://$PGUSER@$PGHOST:$PGPORT/resourced-master-ts-metrics?sslmode=$PGSSLMODE -path $CURRENT_DIR/../../migrations/pg/ts-metrics $@

migrate -url postgres://$PGUSER@$PGHOST:$PGPORT/resourced-master-ts-logs?sslmode=$PGSSLMODE -path $CURRENT_DIR/../../migrations/pg/ts-logs $@

migrate -url postgres://$PGUSER@$PGHOST:$PGPORT/resourced-master-ts-checks?sslmode=$PGSSLMODE -path $CURRENT_DIR/../../migrations/pg/ts-checks $@
