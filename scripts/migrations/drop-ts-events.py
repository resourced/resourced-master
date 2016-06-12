#!/usr/bin/env python

#
# Documentation:
# Purpose: This script helps generate migration file for ts-events child tables.
# Usage: ./scripts/migrations/drop-ts-events.py 2017 > ./migrations/core/0034_add-ts-events-2017.down.sql
# Arguments:
# 1.   The year. Example: 2017
#
# Examples:
# ./scripts/migrations/drop-ts-events.py 2017 > ./migrations/core/0034_add-ts-events-2017.down.sql
# ./scripts/migrations/drop-ts-events.py 2017 > ./migrations/ts-events/0034_add-ts-events-2017.down.sql

import os
import sys

if __name__ == '__main__':
	template_up = os.path.join(os.path.dirname(os.path.realpath(__file__)), '..', '..', 'migrations', 'core', '0003_add-ts-events-2016.down.sql')

	content_up = open(template_up, 'r').read()

	print content_up.replace('2016', sys.argv[1])
