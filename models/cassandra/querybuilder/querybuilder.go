// Package querybuilder parses ResourceD query and turn it into Cassandra + Lucene query.
package querybuilder

import (
	"fmt"
	"strings"

	"github.com/resourced/resourced-master/libstring"
)

// Parse parses ResourceD query and turns it into Cassandra + Lucene query.
func Parse(input string) string {
	return parseAnd(input)
}

// parseAnd parses and conjunctive operator.
func parseAnd(input string) string {
	pgQueryParts := make([]string, 0)

	// normalize variation of AND
	input = strings.Replace(input, " AND ", " and ", -1)

	statements := strings.Split(input, " and ")
	for _, statement := range statements {
		pgStatement := parseStatement(statement)
		if pgStatement != "" {
			pgQueryParts = append(pgQueryParts, pgStatement)
		}
	}

	if len(pgQueryParts) > 1 {
		return fmt.Sprintf(`{type: "boolean", must: [%v]}`, strings.Join(pgQueryParts, ","))

	} else if len(pgQueryParts) == 1 {
		return pgQueryParts[0]
	}

	return ""
}

// parseStringField parses ResourceD "string" query and turns it into lucene search condition.
func parseStringField(statement, field, operator string) string {
	parts := strings.Split(statement, operator)

	data := parts[len(parts)-1]
	data = strings.TrimSpace(data)
	data = libstring.StripChars(data, `"'`)

	if operator == "contains" {
		dataSlice := strings.Split(data, ",")

		for i, datum := range dataSlice {
			dataSlice[i] = `"` + strings.TrimSpace(datum) + `"`
		}

		return fmt.Sprintf(`{type: "%v", field: "%v", values: "%v"}`, operator, field, strings.Join(dataSlice, `,`))
	}

	return fmt.Sprintf(`{type: "%v", field: "%v", value: "%v"}`, operator, field, data)
}

// parseFullTextSearchField parses ResourceD "search" query and turns it into postgres statement.
// Operator is "search" in this context.
func parseFullTextSearchField(statement, field, operator string) string {
	parts := strings.Split(statement, operator)

	searchQuery := parts[len(parts)-1]
	searchQuery = strings.TrimSpace(searchQuery)
	searchQuery = libstring.StripChars(searchQuery, `"'`)

	return fmt.Sprintf(`{type: "phrase", field: "logline", value: "%v", slop: 1}`, searchQuery)
}

// parseStatement parses ResourceD statement and turns it into postgres statement.
func parseStatement(statement string) string {
	statement = strings.TrimSpace(statement)

	// Querying tags.
	// There can only be 1 operator for tags: "="
	if strings.HasPrefix(statement, "Tags") || strings.HasPrefix(statement, "tags") {
		if strings.Contains(statement, "=") {
			parts := strings.Split(statement, "=")

			field := strings.Replace(parts[0], ".", "$", 1)

			data := parts[len(parts)-1]
			data = strings.TrimSpace(data)
			data = libstring.StripChars(data, `"'`)

			return fmt.Sprintf(`{type: "%v", field: "%v", value: "%v"}`, "match", field, data)
		}
	}

	// Querying hostname.
	// Operators:
	// "="        : Exact match.
	// "^"        : Starts with, case sensitive.
	// "~"        : Matches regular expression, case sensitive.
	// "contains" : Contains the following values.
	// "wildcard" : Perform wildcard search.
	if strings.HasPrefix(statement, "Hostname") || strings.HasPrefix(statement, "hostname") {
		if strings.Contains(statement, "=") {
			parts := strings.Split(statement, "=")

			data := parts[len(parts)-1]
			data = strings.TrimSpace(data)
			data = libstring.StripChars(data, `"'`)

			return fmt.Sprintf(`{type: "%v", field: "%v", value: "%v"}`, "match", "hostname", data)

		} else if strings.Contains(statement, "contains") {
			return parseStringField(statement, "hostname", "contains")

		} else if strings.Contains(statement, "wildcard") {
			return parseStringField(statement, "hostname", "wildcard")

		} else if strings.Contains(statement, "^") {
			return parseStringField(statement, "hostname", "prefix")

		} else if strings.Contains(statement, "~") {
			return parseStringField(statement, "hostname", "regexp")
		}
	}

	// Querying filename.
	// Operators:
	// "="        : Exact match.
	// "^"        : Starts with, case sensitive.
	// "~"        : Matches regular expression, case sensitive.
	// "contains" : Contains the following values.
	// "wildcard" : Perform wildcard search.
	if strings.HasPrefix(statement, "Filename") || strings.HasPrefix(statement, "filename") {
		if strings.Contains(statement, "=") {
			parts := strings.Split(statement, "=")

			data := parts[len(parts)-1]
			data = strings.TrimSpace(data)
			data = libstring.StripChars(data, `"'`)

			return fmt.Sprintf(`{type: "%v", field: "%v", value: "%v"}`, "match", "filename", data)

		} else if strings.Contains(statement, "contains") {
			return parseStringField(statement, "filename", "contains")

		} else if strings.Contains(statement, "wildcard") {
			return parseStringField(statement, "filename", "wildcard")

		} else if strings.Contains(statement, "^") {
			return parseStringField(statement, "filename", "prefix")

		} else if strings.Contains(statement, "~") {
			return parseStringField(statement, "filename", "regexp")
		}
	}

	// Querying logline.
	// Operators:
	// "search" : Full text search.
	if strings.HasPrefix(statement, "Logline") || strings.HasPrefix(statement, "logline") {
		if strings.Contains(statement, "search") {
			return parseFullTextSearchField(statement, "logline", "search")
		}
	}

	return ""
}
