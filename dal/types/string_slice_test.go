package types

import (
	"testing"
)

func TestStringSliceScan(t *testing.T) {
	var slice StringSlice

	err := slice.Scan([]byte(`{"12",45,"abc,\\\"d\\ef\\\\"}`))

	if err != nil {
		t.Errorf("Could not scan array, %v", err)
		return
	}

	if slice[0] != "12" || slice[2] != `abc,\"d\ef\\` {
		t.Errorf("Did not get expected slice contents")
	}
}

func TestStringSliceDbValue(t *testing.T) {
	slice := StringSlice([]string{`as"f\df`, "43", "}adsf"})

	val, err := slice.Value()
	if err != nil {
		t.Errorf("Could not convert to db string")
	}

	if str, ok := val.(string); ok {
		if `{"as\"f\\\df","43","}adsf"}` != str {
			t.Errorf("db value expecting %s got %s", `{"as\"f\\\df","43","}adsf"}`, str)
		}
	} else {
		t.Errorf("Could not convert %v to string for comparison", val)
	}
}
