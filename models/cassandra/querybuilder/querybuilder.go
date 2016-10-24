// Package querybuilder parses ResourceD query and turn it into Cassandra + Lucene query.
package querybuilder

import (
	"fmt"
	"strings"

	"github.com/resourced/resourced-master/libstring"
)

// Parse parses ResourceD query and turns it into Cassandra + Lucene query.
func Parse(input string, skipFields []string) string {
	return parseAnd(input, skipFields)
}

// parseAnd parses and conjunctive operator.
func parseAnd(input string, skipFields []string) string {
	pgQueryParts := make([]string, 0)

	// normalize variation of AND
	input = strings.Replace(input, " AND ", " and ", -1)

	statements := strings.Split(input, " and ")
	for _, statement := range statements {
		pgStatement := parseStatement(statement, skipFields)
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

// parseStatement parses ResourceD statement and turns it into postgres statement.
func parseStatement(statement string, skipFields []string) string {
	statement = strings.TrimSpace(statement)

	// Querying tags.
	// There can only be 1 operator for tags: "="
	if strings.HasPrefix(statement, "Tags") || strings.HasPrefix(statement, "tags") {
		if strings.Contains(statement, "=") {
			parts := strings.Split(statement, "=")

			data := parts[len(parts)-1]
			data = strings.TrimSpace(data)
			data = libstring.StripChars(data, `"'`)

			fieldTags := strings.Replace(parts[0], ".", "$", 1)
			fieldTags = strings.TrimSpace(fieldTags)

			foundMasterTagsToSkip := false
			for _, skipField := range skipFields {
				if skipField == "master_tags" {
					foundMasterTagsToSkip = true
					break
				}
			}

			// Skip master_tags
			if foundMasterTagsToSkip {
				return fmt.Sprintf(`{type: "match", field: "%v", value: "%v"}`, fieldTags, data)
			}

			// Not skipping master_tags
			fieldMasterTags := strings.Replace(fieldTags, "tags$", "master_tags$", 1)

			return fmt.Sprintf(`{type: "boolean", should: [{type: "match", field: "%v", value: "%v"},{type: "match", field: "%v", value: "%v"}]}`, fieldTags, data, fieldMasterTags, data)
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
			parts := strings.Split(statement, "contains")

			data := parts[len(parts)-1]
			data = strings.TrimSpace(data)
			data = libstring.StripChars(data, `"'`)

			dataSlice := strings.Split(data, ",")
			for i, datum := range dataSlice {
				dataSlice[i] = `"` + strings.TrimSpace(datum) + `"`
			}

			return fmt.Sprintf(`{type: "%v", field: "%v", values: "%v"}`, "contains", "hostname", strings.Join(dataSlice, `,`))

		} else if strings.Contains(statement, "wildcard") {
			parts := strings.Split(statement, "wildcard")

			data := parts[len(parts)-1]
			data = strings.TrimSpace(data)
			data = libstring.StripChars(data, `"'`)

			return fmt.Sprintf(`{type: "%v", field: "%v", value: "%v"}`, "wildcard", "hostname", data)

		} else if strings.Contains(statement, "^") {
			parts := strings.Split(statement, "^")

			data := parts[len(parts)-1]
			data = strings.TrimSpace(data)
			data = libstring.StripChars(data, `"'`)

			return fmt.Sprintf(`{type: "%v", field: "%v", value: "%v"}`, "prefix", "hostname", data)

		} else if strings.Contains(statement, "~") {
			parts := strings.Split(statement, "~")

			data := parts[len(parts)-1]
			data = strings.TrimSpace(data)
			data = libstring.StripChars(data, `"'`)

			return fmt.Sprintf(`{type: "%v", field: "%v", value: "%v"}`, "regexp", "hostname", data)
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
			parts := strings.Split(statement, "contains")

			data := parts[len(parts)-1]
			data = strings.TrimSpace(data)
			data = libstring.StripChars(data, `"'`)

			dataSlice := strings.Split(data, ",")
			for i, datum := range dataSlice {
				dataSlice[i] = `"` + strings.TrimSpace(datum) + `"`
			}

			return fmt.Sprintf(`{type: "%v", field: "%v", values: "%v"}`, "contains", "filename", strings.Join(dataSlice, `,`))

		} else if strings.Contains(statement, "wildcard") {
			parts := strings.Split(statement, "wildcard")

			data := parts[len(parts)-1]
			data = strings.TrimSpace(data)
			data = libstring.StripChars(data, `"'`)

			return fmt.Sprintf(`{type: "%v", field: "%v", value: "%v"}`, "wildcard", "filename", data)

		} else if strings.Contains(statement, "^") {
			parts := strings.Split(statement, "^")

			data := parts[len(parts)-1]
			data = strings.TrimSpace(data)
			data = libstring.StripChars(data, `"'`)

			return fmt.Sprintf(`{type: "%v", field: "%v", value: "%v"}`, "prefix", "filename", data)

		} else if strings.Contains(statement, "~") {
			parts := strings.Split(statement, "~")

			data := parts[len(parts)-1]
			data = strings.TrimSpace(data)
			data = libstring.StripChars(data, `"'`)

			return fmt.Sprintf(`{type: "%v", field: "%v", value: "%v"}`, "regexp", "filename", data)
		}
	}

	// Querying logline.
	// Operators:
	// "search" : Full text search.
	if strings.HasPrefix(statement, "Logline") || strings.HasPrefix(statement, "logline") {
		if strings.Contains(statement, "search") {
			parts := strings.Split(statement, "search")

			searchQuery := parts[len(parts)-1]
			searchQuery = strings.TrimSpace(searchQuery)
			searchQuery = libstring.StripChars(searchQuery, `"'`)

			return fmt.Sprintf(`{type: "phrase", field: "logline", value: "%v", slop: 1}`, searchQuery)
		}
	}

	return ""
}
