#!/usr/bin/env python

#
# Documentation:
# Purpose: This script helps generate migration file for timeseries child tables.
# Usage: ./scripts/migrations/create-ts-daily.py ts_checks 2016 check_id created > ./migrations/core/0032_add-ts-checks-2016.up.sql
# Arguments:
# 1.   Parent table name. Example: ts_checks
# 2.   The year. Example: 2016
# 3... Columns to create composite indexes. The last column is expected to be TIMESTAMP column where inheritance is based on.
#
# Examples:
# ./scripts/migrations/create-ts-daily.py ts_executor_logs 2016 created > ./migrations/core/0026_add-ts-executor-logs-2016.up.sql
# ./scripts/migrations/create-ts-daily.py ts_executor_logs 2016 created > ./migrations/ts-logs/0026_add-ts-executor-logs-2016.up.sql
# ./scripts/migrations/create-ts-daily.py ts_logs 2016 created > ./migrations/core/0027_add-ts-logs-2016.up.sql
# ./scripts/migrations/create-ts-daily.py ts_logs 2016 created > ./migrations/ts-logs/0027_add-ts-logs-2016.up.sql
# ./scripts/migrations/create-ts-daily.py ts_checks 2016 check_id created > ./migrations/core/0032_add-ts-checks-2016.up.sql
# ./scripts/migrations/create-ts-daily.py ts_checks 2016 check_id created > ./migrations/ts-checks/0032_add-ts-checks-2016.up.sql

import sys
import calendar
from string import Template

def create_table_by_day(table_name, table_ts_column, year, month, index):
	month_range = calendar.monthrange(year, month)
	if index > month_range[1]:
		return ""

	next_year = year
	next_month = month
	next_index = index + 1

	# Check edge cases
	if next_index > month_range[1]:
		next_index = 1
		next_month = next_month + 1

		if next_month > 12:
			next_month = 1
			next_year = year + 1

	padded_month = "%02d" % (month)
	padded_next_month = "%02d" % (next_month)
	padded_index = "%02d" % (index)
	padded_next_index = "%02d" % (next_index)

	table_name_with_suffix = "%s_%s_%s_%s" % (table_name, year, padded_month, padded_index)

	t = Template("create table $table_name_with_suffix (check ($table_ts_column >= TIMESTAMP '$year-$padded_month-$padded_index 00:00:00-00' and $table_ts_column < TIMESTAMP '$next_year-$padded_next_month-$padded_next_index 00:00:00-00')) inherits ($table_name);")
	return t.substitute(
		table_name_with_suffix=table_name_with_suffix,
		table_name=table_name,
		table_ts_column=table_ts_column,
		year=year,
		next_year=next_year,
		padded_month=padded_month,
		padded_next_month=padded_next_month,
		padded_index=padded_index,
		padded_next_index=padded_next_index
	)

def create_brin_index_by_day(table_name, year, month, index, *columns):
	month_range = calendar.monthrange(year, month)
	if index > month_range[1]:
		return ""

	padded_month = "%02d" % (month)
	padded_index = "%02d" % (index)

	index_name_with_suffix = "idx_%s_%s_%s_%s_%s" % (table_name, year, padded_month, padded_index, '_'.join(columns))
	table_name_with_suffix = "%s_%s_%s_%s" % (table_name, year, padded_month, padded_index)

	t = Template("create index $index_name_with_suffix on $table_name_with_suffix using brin ($comma_sep_columns);")
	return t.substitute(
		index_name_with_suffix=index_name_with_suffix,
		table_name_with_suffix=table_name_with_suffix,
		table_name=table_name,
		year=year,
		comma_sep_columns=','.join(columns)
	)

def create_ts_logs_fulltext_index_by_day(year, month, index):
	month_range = calendar.monthrange(year, month)
	if index > month_range[1]:
		return ""

	padded_month = "%02d" % (month)
	padded_index = "%02d" % (index)
	table_name = "ts_logs"
	columns = ["cluster_id", "created", "hostname", "tags", "to_tsvector('english', logline)"]

	index_name_with_suffix = "idx_%s_%s_%s_%s_%s" % (table_name, year, padded_month, padded_index, '_'.join(columns[:-1]))
	table_name_with_suffix = "%s_%s_%s_%s" % (table_name, year, padded_month, padded_index)

	t = Template("create index $index_name_with_suffix on $table_name_with_suffix using gin($comma_sep_columns);")
	return t.substitute(
		index_name_with_suffix=index_name_with_suffix,
		table_name_with_suffix=table_name_with_suffix,
		table_name=table_name,
		year=year,
		comma_sep_columns=','.join(columns)
	)

def create_function_on_insert_if_clause(table_name, table_ts_column, year, month, index):
	month_range = calendar.monthrange(year, month)
	if index > month_range[1]:
		return ""

	next_year = year
	next_month = month
	next_index = index + 1

	# Check edge cases
	if next_index > month_range[1]:
		next_index = 1
		next_month = next_month + 1

		if next_month > 12:
			next_month = 1
			next_year = year + 1

	padded_month = "%02d" % (month)
	padded_next_month = "%02d" % (next_month)
	padded_index = "%02d" % (index)
	padded_next_index = "%02d" % (next_index)

	table_name_with_suffix = "%s_%s_%s_%s" % (table_name, year, padded_month, padded_index)
	if_or_elsif = "if" if (index == 1 and month == 1) else "elsif"

	t = Template("""    $if_or_elsif ( new.$table_ts_column >= TIMESTAMP '$year-$padded_month-$padded_index 00:00:00-00' and new.$table_ts_column < TIMESTAMP '$next_year-$padded_next_month-$padded_next_index 00:00:00-00') then
        insert into $table_name_with_suffix values (new.*);"""
    )

	return t.substitute(
		if_or_elsif=if_or_elsif,
		table_ts_column=table_ts_column,
		table_name_with_suffix=table_name_with_suffix,
		year=year,
		next_year=next_year,
		padded_month=padded_month,
		padded_next_month=padded_next_month,
		padded_index=padded_index,
		padded_next_index=padded_next_index
	)

def create_function_on_insert(table_name, table_ts_column, year):
	function_name = "on_%s_insert_%s" % (table_name, year)
	trigger_name = "trigger_%s" % (function_name)
	if_clauses_list = filter(bool, [create_function_on_insert_if_clause(table_name, table_ts_column, year, month, index) for month in range(1,13) for index in range(1,32)])
	if_clauses = "\n".join(if_clauses_list)
	double_dollar = "$$"

	t = Template("""create or replace function $function_name() returns trigger as $double_dollar
begin
$if_clauses
    end if;

    return null;
end;
$double_dollar language plpgsql;

create trigger $trigger_name
    before insert on $table_name
    for each row execute procedure $function_name();
""")

	return t.substitute(
		function_name=function_name,
		trigger_name=trigger_name,
		table_name=table_name,
		if_clauses=if_clauses,
		double_dollar=double_dollar
	)

def create_migration(table_name, year, *columns):
	create_tables_list = filter(bool, [create_table_by_day(table_name, columns[-1], year, month, index) for month in range(1,13) for index in range(1,32)])

	if table_name == "ts_logs":
		create_indexes_list = filter(bool, [create_ts_logs_fulltext_index_by_day(year, month, index) for month in range(1,13) for index in range(1,32)])
	else:
		create_indexes_list = filter(bool, [create_brin_index_by_day(table_name, year, month, index, *(['cluster_id'] + list(columns))) for month in range(1,13) for index in range(1,32)])

	create_tables = "\n".join(create_tables_list)
	create_indexes = "\n".join(create_indexes_list)
	trigger_on_insert = create_function_on_insert(table_name, columns[-1], year)

	t = Template("""$create_tables

$create_indexes

$trigger_on_insert
""")

	return t.substitute(
		create_tables=create_tables,
		create_indexes=create_indexes,
		trigger_on_insert=trigger_on_insert
	)


if __name__ == '__main__':
	print(create_migration(sys.argv[1], int(sys.argv[2]), *sys.argv[3:]))
