package libcollection

import (
	"encoding/json"
	"testing"
)

func TestFlattenMap(t *testing.T) {
	inputJson := `{"/free": {"Swap": {"Free": 0, "Used": 0, "Total": 0}}}`
	input := make(map[string]interface{})

	json.Unmarshal([]byte(inputJson), &input)

	output := make(map[string]interface{})

	FlattenMap(input, ".", "", &output)

	value, ok := output["/free.Swap.Free"]
	if !ok {
		t.Errorf(`output["/free.Swap.Free"] should return values.`)
	}
	if value.(float64) != 0 {
		t.Errorf(`output["/free.Swap.Free"] should return values.`)
	}
}
