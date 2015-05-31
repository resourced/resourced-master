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

func TestParseNameExact(t *testing.T) {
	toBeTested := []string{
		`Name = "Awesome Sauce"`,
		`Name="Awesome Sauce"`,
		`name = "Awesome Sauce"`,
	}

	for _, testString := range toBeTested {
		output := Parse(testString)
		if output != `name = 'Awesome Sauce'` {
			t.Errorf("Failed to generate name query. Output: %v", output)
		}
	}
}

func TestParseNameStartsWith(t *testing.T) {
	toBeTested := []string{
		`Name ~^ "brotato"`,
		`Name~^"brotato"`,
		`name ~^ "brotato"`,
	}

	for _, testString := range toBeTested {
		output := Parse(testString)
		if output != `name LIKE 'brotato%'` {
			t.Errorf("Failed to generate name query. Output: %v", output)
		}
	}
}

func TestParseNameDoesNotMatchRegexCaseInsensitive(t *testing.T) {
	toBeTested := []string{
		`Name !~* "brotato"`,
		`Name!~*"brotato"`,
		`name !~* "brotato"`,
	}

	for _, testString := range toBeTested {
		output := Parse(testString)
		if output != `name !~* 'brotato'` {
			t.Errorf("Failed to generate name query. Output: %v", output)
		}
	}
}

func TestParseNameDoesNotMatchRegexCaseSensitive(t *testing.T) {
	toBeTested := []string{
		`Name !~ "brotato"`,
		`Name!~"brotato"`,
		`name !~ "brotato"`,
	}

	for _, testString := range toBeTested {
		output := Parse(testString)
		if output != `name !~ 'brotato'` {
			t.Errorf("Failed to generate name query. Output: %v", output)
		}
	}
}

func TestParseNameMatchRegexCaseInsensitive(t *testing.T) {
	toBeTested := []string{
		`Name ~* "brotato"`,
		`Name~*"brotato"`,
		`name ~* "brotato"`,
	}

	for _, testString := range toBeTested {
		output := Parse(testString)
		if output != `name ~* 'brotato'` {
			t.Errorf("Failed to generate name query. Output: %v", output)
		}
	}
}

func TestParseNameMatchRegexCaseSensitive(t *testing.T) {
	toBeTested := []string{
		`Name ~ "brotato"`,
		`Name~"brotato"`,
		`name ~ "brotato"`,
	}

	for _, testString := range toBeTested {
		output := Parse(testString)
		if output != `name ~ 'brotato'` {
			t.Errorf("Failed to generate name query. Output: %v", output)
		}
	}
}

func TestParseJsonTraversal(t *testing.T) {
	// Example JSON:
	// {"/free": {"Swap": {"Free": 0, "Used": 0, "Total": 0}, "Memory": {"Free": 1346609152, "Used": 7243325440, "Total": 8589934592, "ActualFree": 3666075648, "ActualUsed": 4923858944}}}

	toBeTested := `/free.Memory.Free > 10000000`

	output := Parse(toBeTested)
	if output != `data #>> '{/free,Memory,Free}' > '10000000'` {
		t.Errorf("Failed to generate name query. Output: %v", output)
	}
}

func TestParseAnd(t *testing.T) {
	toBeTested := `tags = ["aaa","bbb","ccc"] AND Name~^"brotato" AND /free.Memory.Free > 10000000`

	output := Parse(toBeTested)
	if output != `tags ?& array["aaa","bbb","ccc"] AND name LIKE 'brotato%' AND data #>> '{/free,Memory,Free}' > '10000000'` {
		t.Errorf("Failed to generate name query. Output: %v", output)
	}
}
