package querybuilder

import (
	"fmt"
	"github.com/resourced/resourced-master/libstring"
	"strings"
)

func Parse(input string) string {
	pgQueryParts := make([]string, 0)

	statements := strings.Split(input, " AND ")
	for _, statement := range statements {
		statement = strings.TrimSpace(statement)

		// Querying tags.
		// There can only be 1 operator for tags: "="
		if strings.HasPrefix(statement, "Tags") || strings.HasPrefix(statement, "tags") {
			parts := strings.Split(statement, "=")

			arrayOfTagsInString := parts[len(parts)-1]
			arrayOfTagsInString = strings.TrimSpace(arrayOfTagsInString)
			arrayOfTagsInString = libstring.StripChars(arrayOfTagsInString, "[]")

			query := fmt.Sprintf("tags ?& array[%v]", arrayOfTagsInString)

			pgQueryParts = append(pgQueryParts, query)
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

				query := fmt.Sprintf("name = %v", name)
				pgQueryParts = append(pgQueryParts, query)

			} else if strings.Contains(statement, "~^") {
				parts := strings.Split(statement, "~^")

				name := parts[len(parts)-1]
				name = strings.TrimSpace(name)
				name = libstring.StripChars(name, `"'`)

				query := `name LIKE "` + name + `%"`
				pgQueryParts = append(pgQueryParts, query)
			}
		}
	}

	if len(pgQueryParts) > 1 {
		return strings.Join(pgQueryParts, " AND ")
	} else if len(pgQueryParts) == 1 {
		return pgQueryParts[0]
	}

	return ""
}
