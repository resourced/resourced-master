// Package querybuilder parses ResourceD query and turn it into PostgreSQL query.
package querybuilder

import (
	"fmt"
	"strings"

	"github.com/resourced/resourced-master/libstring"
)

// Parse parses ResourceD query and turns it into postgres query.
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
		return strings.Join(pgQueryParts, " and ")
	} else if len(pgQueryParts) == 1 {
		return pgQueryParts[0]
	}

	return ""
}

// parseStringField parses ResourceD "string" query and turns it into postgres statement.
func parseStringField(statement, field, operator string) string {
	parts := strings.Split(statement, operator)

	data := parts[len(parts)-1]
	data = strings.TrimSpace(data)
	data = libstring.StripChars(data, `"'`)

	return fmt.Sprintf("%v %v '%v'", field, operator, data)
}

// parseFullTextSearchField parses ResourceD "search" query and turns it into postgres statement.
// Operator is "search" in this context.
func parseFullTextSearchField(statement, field, operator string) string {
	parts := strings.Split(statement, operator)

	searchQuery := parts[len(parts)-1]
	searchQuery = strings.TrimSpace(searchQuery)
	searchQuery = libstring.StripChars(searchQuery, `"'`)

	// Split search query into multiple soon-to-be-tsquery-function-argument (+ boolean operators)
	searchQueryChunks := strings.Fields(searchQuery)

	tsQueryChunksWithBoolOperators := make([]string, 0)
	tsQueryChunk := ""

	for i, chunk := range searchQueryChunks {
		// If chunk is not boolean operator, keep on appending to the same chunk.
		if chunk != "||" && chunk != "&&" {
			tsQueryChunk = tsQueryChunk + " " + chunk

		} else {
			tsQueryChunksWithBoolOperators = append(tsQueryChunksWithBoolOperators, strings.TrimSpace(tsQueryChunk))
			tsQueryChunk = ""

			// append boolean operator as well
			tsQueryChunksWithBoolOperators = append(tsQueryChunksWithBoolOperators, strings.TrimSpace(chunk))
		}

		// If reached the end
		if i == len(searchQueryChunks)-1 {
			if strings.TrimSpace(tsQueryChunk) != "" {
				tsQueryChunksWithBoolOperators = append(tsQueryChunksWithBoolOperators, strings.TrimSpace(tsQueryChunk))
			}
		}
	}

	// Now, let's build series of tsquery functions
	tsQuerySlice := make([]string, 0)

	for _, chunk := range tsQueryChunksWithBoolOperators {
		if chunk == "||" || chunk == "&&" {
			tsQuerySlice = append(tsQuerySlice, chunk)

		} else if strings.Contains(chunk, "|") || strings.Contains(chunk, "&") {
			tsQuerySlice = append(tsQuerySlice, fmt.Sprintf("to_tsquery('english', '%v')", chunk))

		} else {
			tsQuerySlice = append(tsQuerySlice, fmt.Sprintf("plainto_tsquery('english', '%v')", chunk))
		}
	}

	tsQueries := strings.Join(tsQuerySlice, " ")

	return fmt.Sprintf(`to_tsvector('english', regexp_replace(%v, '[^\w]+', ' ', 'gi')) || to_tsvector('english', %v) @@ (%v)`, field, field, tsQueries)
}

// parseStatement parses ResourceD statement and turns it into postgres statement.
func parseStatement(statement string) string {
	statement = strings.TrimSpace(statement)

	// Querying tags.
	// There can only be 1 operator for tags: "="
	if strings.HasPrefix(statement, "Tags") || strings.HasPrefix(statement, "tags") {
		operator := ""

		if strings.Contains(statement, "=") {
			operator = "="
		}

		if operator != "" {
			parts := strings.Split(statement, operator)

			// Remove tags prefix
			parts[0] = strings.Replace(parts[0], "tags.", "", -1)

			pgJsonPath := strings.Replace(parts[0], ".", ",", -1)
			pgJsonPath = strings.TrimSpace(pgJsonPath)

			value := parts[len(parts)-1]
			value = strings.TrimSpace(value)
			value = libstring.StripChars(value, `"'`)

			return fmt.Sprintf("tags #>> '{%v}' %v '%v'", pgJsonPath, operator, value)
		}
	}

	// Querying hostname.
	// Operators:
	// "="   : Exact match.
	// "!~*" : Does not match regular expression, case insensitive.
	// "!~"  : Does not match regular expression, case sensitive.
	// "~*"  : Matches regular expression, case insensitive.
	// "~^"  : Starts with, case sensitive.
	// "~"   : Matches regular expression, case sensitive.
	if strings.HasPrefix(statement, "Hostname") || strings.HasPrefix(statement, "hostname") {
		if strings.Contains(statement, "=") {
			return parseStringField(statement, "hostname", "=")

		} else if strings.Contains(statement, "!~*") {
			return parseStringField(statement, "hostname", "!~*")

		} else if strings.Contains(statement, "!~") {
			return parseStringField(statement, "hostname", "!~")

		} else if strings.Contains(statement, "~*") {
			return parseStringField(statement, "hostname", "~*")

		} else if strings.Contains(statement, "~^") {
			parts := strings.Split(statement, "~^")

			hostname := parts[len(parts)-1]
			hostname = strings.TrimSpace(hostname)
			hostname = libstring.StripChars(hostname, `"'`)

			return `hostname LIKE '` + hostname + `%'`

		} else if strings.Contains(statement, "~") {
			return parseStringField(statement, "hostname", "~")
		}
	}

	// Querying filename.
	// Operators:
	// "="   : Exact match.
	// "!~*" : Does not match regular expression, case insensitive.
	// "!~"  : Does not match regular expression, case sensitive.
	// "~*"  : Matches regular expression, case insensitive.
	// "~^"  : Starts with, case sensitive.
	// "~"   : Matches regular expression, case sensitive.
	if strings.HasPrefix(statement, "Filename") || strings.HasPrefix(statement, "filename") {
		if strings.Contains(statement, "=") {
			return parseStringField(statement, "filename", "=")

		} else if strings.Contains(statement, "!~*") {
			return parseStringField(statement, "filename", "!~*")

		} else if strings.Contains(statement, "!~") {
			return parseStringField(statement, "filename", "!~")

		} else if strings.Contains(statement, "~*") {
			return parseStringField(statement, "filename", "~*")

		} else if strings.Contains(statement, "~^") {
			parts := strings.Split(statement, "~^")

			filename := parts[len(parts)-1]
			filename = strings.TrimSpace(filename)
			filename = libstring.StripChars(filename, `"'`)

			return `filename LIKE '` + filename + `%'`

		} else if strings.Contains(statement, "~") {
			return parseStringField(statement, "filename", "~")
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
