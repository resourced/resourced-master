package querybuilder

import (
	"testing"
)

func TestParseTags(t *testing.T) {
	toBeTested := []string{
		`Tags = ["aaa","bbb","ccc"]`,
		`Tags=["aaa","bbb","ccc"]`,
		`tags = ["aaa","bbb","ccc"]`,
	}

	for _, testString := range toBeTested {
		output := Parse(testString)
		if output != `tags ?& array["aaa","bbb","ccc"]` {
			t.Errorf("Failed to generate tags query. Output: %v", output)
		}
	}
}
