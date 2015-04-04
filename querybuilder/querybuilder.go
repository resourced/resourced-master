// Package querybuilder provides query building functionality.
package querybuilder

import (
	"fmt"
	"github.com/resourced/resourced-master/libstring"
	"strings"
)

// Parse parses ResourceD query and turns it into postgres query.
func Parse(input string) string {
	return parseAnd(input)
}

func parseAnd(input string) string {
	pgQueryParts := make([]string, 0)

	statements := strings.Split(input, " AND ")
	for _, statement := range statements {
		pgStatement := parseStatement(statement)
		if pgStatement != "" {
			pgQueryParts = append(pgQueryParts, pgStatement)
		}
	}

	if len(pgQueryParts) > 1 {
		return strings.Join(pgQueryParts, " AND ")
	} else if len(pgQueryParts) == 1 {
		return pgQueryParts[0]
	}

	return ""
}

// parseStatement parses ResourceD statement and turns it into postgres statement.
func parseStatement(statement string) string {
	statement = strings.TrimSpace(statement)

	// Querying tags.
	// There can only be 1 operator for tags: "="
	if strings.HasPrefix(statement, "Tags") || strings.HasPrefix(statement, "tags") {
		parts := strings.Split(statement, "=")

		arrayOfTagsInString := parts[len(parts)-1]
		arrayOfTagsInString = strings.TrimSpace(arrayOfTagsInString)
		arrayOfTagsInString = libstring.StripChars(arrayOfTagsInString, "[]")

		return fmt.Sprintf("tags ?& array[%v]", arrayOfTagsInString)
	}

	// Querying name.
	// Operators:
	// "="  : Exact match.
	// "~^" : Starts with.
	if strings.HasPrefix(statement, "Name") || strings.HasPrefix(statement, "name") {
		if strings.Contains(statement, "=") {
			parts := strings.Split(statement, "=")

			name := parts[len(parts)-1]
			name = strings.TrimSpace(name)

			return fmt.Sprintf("name = %v", name)

		} else if strings.Contains(statement, "~^") {
			parts := strings.Split(statement, "~^")

			name := parts[len(parts)-1]
			name = strings.TrimSpace(name)
			name = libstring.StripChars(name, `"'`)

			return `name LIKE "` + name + `%"`
		}
	}

	// Querying data.
	// Operators: >=, <=, =, <, >
	// Expected output: data #>> '{/free,Swap,Free}' = '0'
	if strings.HasPrefix(statement, "/") {
		operator := ""

		if strings.Contains(statement, ">=") {
			operator = ">="
		} else if strings.Contains(statement, "<=") {
			operator = "<="
		} else if strings.Contains(statement, "=") {
			operator = "="
		} else if strings.Contains(statement, "<") {
			operator = ">"
		} else if strings.Contains(statement, ">") {
			operator = ">"
		}

		if operator != "" {
			parts := strings.Split(statement, operator)

			pgJsonPath := strings.Replace(parts[0], ".", ",", -1)
			pgJsonPath = strings.TrimSpace(pgJsonPath)

			value := parts[len(parts)-1]
			value = strings.TrimSpace(value)
			value = libstring.StripChars(value, `"'`)

			return fmt.Sprintf("data #>> '{%v}' %v '%v'", pgJsonPath, operator, value)
		}
	}

	return ""
}
