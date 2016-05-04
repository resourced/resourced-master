package libslice

import (
	"strings"
)

func RemoveEmpty(s []string) []string {
	result := make([]string, 0)

	for _, str := range s {
		if strings.TrimSpace(str) != "" {
			result = append(result, str)
		}
	}
	return result
}
