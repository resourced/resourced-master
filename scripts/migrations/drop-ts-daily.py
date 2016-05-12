#!/usr/bin/env python

import sys
import calendar
from string import Template

def drop_table_by_day(table_name, table_ts_column, year, month, index):
	month_range = calendar.monthrange(year, month)
	if index > month_range[1]:
		return ""

	padded_month = "%02d" % (month)
	padded_index = "%02d" % (index)

	return "DROP TABLE IF EXISTS %s_%s_%s_%s CASCADE;" % (table_name, year, padded_month, padded_index)

def drop_function_on_insert(table_name, table_ts_column, year):
	function_name = "on_%s_insert_%s" % (table_name, year)
	return "DROP FUNCTION IF EXISTS %s() CASCADE;" % function_name

def create_migration(table_name, table_ts_column, year):
	drop_tables_list = filter(bool, [drop_table_by_day(table_name, table_ts_column, year, month, index) for month in range(1,13) for index in range(1,32)])
	drop_tables = "\n".join(drop_tables_list)
	drop_func = drop_function_on_insert(table_name, table_ts_column, year)

	t = Template("""$drop_tables

$drop_func
""")

	return t.substitute(
		drop_tables=drop_tables,
		drop_func=drop_func
	)


if __name__ == '__main__':
	print(create_migration(sys.argv[1], sys.argv[2], int(sys.argv[3])))
