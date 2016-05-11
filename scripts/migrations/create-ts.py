#!/usr/bin/env python

import calendar
from string import Template

def create_table_by_month(table_name, table_ts_column, year, index):
	padded_index = "%02d" % (index)
	padded_index_plus_one = "%02d" % (index + 1)
	table_name_with_suffix = "%s_m%s_%s" % (table_name, index, year)

	if index == 12:
		padded_index_plus_one = "%02d" % (1)
		year = year + 1

	t = Template("create table $table_name_with_suffix (check ($table_ts_column >= TIMESTAMP '$year-$padded_index-01 00:00:00-00' and $table_ts_column < TIMESTAMP '$year-$padded_index_plus_one-01 00:00:00-00')) inherits ($table_name);")
	return t.substitute(
		table_name_with_suffix=table_name_with_suffix,
		table_name=table_name,
		table_ts_column=table_ts_column,
		year=year,
		padded_index=padded_index,
		padded_index_plus_one=padded_index_plus_one
	)

def create_table_by_day(table_name, table_ts_column, year, month, index):
	month_range = calendar.monthrange(year, month)
	if index > month_range[1]:
		return ""

	table_name_with_suffix = "%s_m%s_d%s_%s" % (table_name, month, index, year)

	next_month = month
	next_index = index + 1

	# Check edge cases
	if next_index > month_range[1]:
		next_index = 1
		next_month = next_month + 1

		if next_month > 12:
			next_month = 1
			year = year + 1

		padded_month = "%02d" % (month)


	padded_month = "%02d" % (month)
	padded_next_month = "%02d" % (next_month)
	padded_index = "%02d" % (index)
	padded_next_index = "%02d" % (next_index)

	t = Template("create table $table_name_with_suffix (check ($table_ts_column >= TIMESTAMP '$year-$padded_month-$padded_index 00:00:00-00' and $table_ts_column < TIMESTAMP '$year-$padded_next_month-$padded_next_index 00:00:00-00')) inherits ($table_name);")
	return t.substitute(
		table_name_with_suffix=table_name_with_suffix,
		table_name=table_name,
		table_ts_column=table_ts_column,
		year=year,
		padded_month=padded_month,
		padded_next_month=padded_next_month,
		padded_index=padded_index,
		padded_next_index=padded_next_index
	)

for i in range(12):
	# print(create_table_by_month('ts_metrics', 'created', 2016, i + 1))

	for j in range(31):
		print(create_table_by_day('ts_metrics', 'created', 2016, i + 1, j + 1))
