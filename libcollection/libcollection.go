// Package libstring provides slice/map related library functions.
package libcollection

import (
	"strconv"
)

// FlattenMap flattens nested map.
func FlattenMap(inputMap map[string]interface{}, separator string, lkey string, flattened *map[string]interface{}) {
	for rkey, value := range inputMap {
		key := lkey + rkey
		if _, ok := value.(string); ok {
			(*flattened)[key] = value.(string)
		} else if _, ok := value.(float64); ok {
			(*flattened)[key] = value.(float64)
		} else if _, ok := value.(bool); ok {
			(*flattened)[key] = value.(bool)
		} else if _, ok := value.([]interface{}); ok {
			for i := 0; i < len(value.([]interface{})); i++ {
				if _, ok := value.([]string); ok {
					stringI := string(i)
					(*flattened)[stringI] = value.(string)
					/// think this is wrong
				} else if _, ok := value.([]int); ok {
					stringI := string(i)
					(*flattened)[stringI] = value.(int)
				} else {
					FlattenMap(value.([]interface{})[i].(map[string]interface{}), separator, key+separator+strconv.Itoa(i)+separator, flattened)
				}
			}
		} else {
			FlattenMap(value.(map[string]interface{}), separator, key+separator, flattened)
		}
	}
}
