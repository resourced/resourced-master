#!/usr/bin/env python

#
# Documentation:
# Purpose: This script helps generate migration file for timeseries child tables.
#          Do not use this script for ts_events, use drop-ts-events.py instead.
# Usage: ./scripts/migrations/drop-ts-daily.py ts_checks 2016 > ./migrations/pg/down/0032_ts-checks-2016.sql
# Arguments:
# 1.   Parent table name. Example: ts_checks
# 2.   The year. Example: 2016
#
# Examples:
# ./scripts/migrations/drop-ts-daily.py ts_metrics 2016 > ./migrations/pg/down/0006_ts-metrics-2016.sql
# ./scripts/migrations/drop-ts-daily.py ts_metrics_aggr_15m 2016 > ./migrations/pg/down/0007_ts-metrics-aggr-15m-2016.sql
# ./scripts/migrations/drop-ts-daily.py ts_logs 2016 > ./migrations/pg/down/0027_ts-logs-2016.sql
# ./scripts/migrations/drop-ts-daily.py ts_checks 2016 > ./migrations/pg/down/0032_ts-checks-2016.sql

import sys
import calendar
from string import Template

def drop_table_by_day(table_name, year, month, index):
	month_range = calendar.monthrange(year, month)
	if index > month_range[1]:
		return ""

	padded_month = "%02d" % (month)
	padded_index = "%02d" % (index)

	return "DROP TABLE IF EXISTS %s_%s_%s_%s CASCADE;" % (table_name, year, padded_month, padded_index)

def drop_function_on_insert(table_name, year):
	function_name = "on_%s_insert_%s" % (table_name, year)
	return "DROP FUNCTION IF EXISTS %s() CASCADE;" % function_name

def create_migration(table_name, year):
	drop_tables_list = filter(bool, [drop_table_by_day(table_name, year, month, index) for month in range(1,13) for index in range(1,32)])
	drop_tables = "\n".join(drop_tables_list)
	drop_func = drop_function_on_insert(table_name, year)

	t = Template("""$drop_tables

$drop_func
""")

	return t.substitute(
		drop_tables=drop_tables,
		drop_func=drop_func
	)


if __name__ == '__main__':
	print(create_migration(sys.argv[1], int(sys.argv[2])))
