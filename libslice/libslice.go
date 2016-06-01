// Package libslice provides slice related library functions.
package libslice

import (
	"strings"
)

// RemoveEmpty string in a slice.
func RemoveEmpty(s []string) []string {
	result := make([]string, 0)

	for _, str := range s {
		if strings.TrimSpace(str) != "" {
			result = append(result, str)
		}
	}
	return result
}
