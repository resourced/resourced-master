package querybuilder

import (
	"testing"
)

func TestParseTags(t *testing.T) {
	toBeTested := []string{
		`tags.aaa = bbb`,
	}

	for _, testString := range toBeTested {
		output := Parse(testString, nil)
		expected := `{type: "boolean", should: [{type: "match", field: "tags$aaa", value: "bbb"},{type: "match", field: "master_tags$aaa", value: "bbb"}]}`

		if output != expected {
			t.Errorf("Failed to generate tags query. Output: %v, Expected: %v", output, expected)
		}
	}
}

func TestParseFilenameExact(t *testing.T) {
	toBeTested := []string{
		`Filename = "/var/log/message"`,
		`Filename="/var/log/message"`,
		`filename = "/var/log/message"`,
	}

	for _, testString := range toBeTested {
		output := Parse(testString, nil)
		if output != `{type: "match", field: "filename", value: "/var/log/message"}` {
			t.Errorf("Failed to generate filename query. Output: %v", output)
		}
	}
}

func TestParseHostnameExact(t *testing.T) {
	toBeTested := []string{
		`Hostname = "Awesome Sauce"`,
		`Hostname="Awesome Sauce"`,
		`hostname = "Awesome Sauce"`,
	}

	for _, testString := range toBeTested {
		output := Parse(testString, nil)
		if output != `{type: "match", field: "hostname", value: "Awesome Sauce"}` {
			t.Errorf("Failed to generate hostname query. Output: %v", output)
		}
	}
}

func TestParseHostnameStartsWith(t *testing.T) {
	toBeTested := []string{
		`Hostname ~^ "brotato"`,
		`Hostname~^"brotato"`,
		`hostname ~^ "brotato"`,
	}

	for _, testString := range toBeTested {
		output := Parse(testString, nil)
		if output != `{type: "prefix", field: "hostname", value: "brotato"}` {
			t.Errorf("Failed to generate hostname query. Output: %v", output)
		}
	}
}

func TestParseHostnameMatchRegexCaseSensitive(t *testing.T) {
	toBeTested := []string{
		`Hostname ~ "brotato"`,
		`Hostname~"brotato"`,
		`hostname ~ "brotato"`,
	}

	for _, testString := range toBeTested {
		output := Parse(testString, nil)
		if output != `{type: "regexp", field: "hostname", value: "brotato"}` {
			t.Errorf("Failed to generate hostname query. Output: %v", output)
		}
	}
}

// func TestParseJsonTraversalFloatComparison(t *testing.T) {
// 	// Example JSON:
// 	// {"/free": {"Swap": {"Free": 0, "Used": 0, "Total": 0}, "Memory": {"Free": 1346609152, "Used": 7243325440, "Total": 8589934592, "ActualFree": 3666075648, "ActualUsed": 4923858944}}}

// 	toBeTested := `/free.Memory.Free > 10000000`
// 	output := Parse(toBeTested)
// 	expected := `(data #>> '{/free,Memory.Free}')::float8 > 10000000`

// 	if output != expected {
// 		t.Errorf("Failed to generate data query. Output: %v, Expected: %v", output, expected)
// 	}
// }

// func TestParseJsonTraversalStringComparison(t *testing.T) {
// 	toBeTested := `/Uname.Shell ~ "Darwin"`
// 	output := Parse(toBeTested)
// 	expected := `data #>> '{/Uname,Shell}' ~ 'Darwin'`

// 	if output != expected {
// 		t.Errorf("Failed to generate data query. Output: %v, Expected: %v", output, expected)
// 	}
// }

// func TestParseJsonTraversalEquality(t *testing.T) {
// 	toBeTested := `/Uname.Shell = "Darwin"`
// 	output := Parse(toBeTested)
// 	expected := `data #>> '{/Uname,Shell}' = 'Darwin'`

// 	if output != expected {
// 		t.Errorf("Failed to generate data query. Output: %v, Expected: %v", output, expected)
// 	}

// 	toBeTested = `/free.Memory.Free = 10000000`
// 	output = Parse(toBeTested)
// 	expected = `data #>> '{/free,Memory.Free}' = '10000000'`

// 	if output != expected {
// 		t.Errorf("Failed to generate data query. Output: %v, Expected: %v", output, expected)
// 	}
// }

func TestParseAnd(t *testing.T) {
	toBeTested := `tags.aaa = bbb AND Hostname~^"brotato" AND /free.Memory.Free > 10000000`
	output := Parse(toBeTested, nil)
	expected := `{type: "boolean", must: [{type: "boolean", should: [{type: "match", field: "tags$aaa", value: "bbb"},{type: "match", field: "master_tags$aaa", value: "bbb"}]},{type: "prefix", field: "hostname", value: "brotato"}]}`

	if output != expected {
		t.Errorf("Failed to generate mixed of tags,hostname, and data query. Output: %v, Expected: %v", output, expected)
	}
}
