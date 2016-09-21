package querybuilder

import (
	"testing"
)

func TestParseTags(t *testing.T) {
	toBeTested := []string{
		`tags.aaa = bbb`,
	}

	for _, testString := range toBeTested {
		output := Parse(testString)
		if output != `tags #>> '{aaa}' = 'bbb'` {
			t.Errorf("Failed to generate tags query. Output: %v", output)
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
		output := Parse(testString)
		if output != `filename = '/var/log/message'` {
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
		output := Parse(testString)
		if output != `hostname = 'Awesome Sauce'` {
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
		output := Parse(testString)
		if output != `hostname LIKE 'brotato%'` {
			t.Errorf("Failed to generate hostname query. Output: %v", output)
		}
	}
}

func TestParseHostnameDoesNotMatchRegexCaseInsensitive(t *testing.T) {
	toBeTested := []string{
		`Hostname !~* "brotato"`,
		`Hostname!~*"brotato"`,
		`hostname !~* "brotato"`,
	}

	for _, testString := range toBeTested {
		output := Parse(testString)
		if output != `hostname !~* 'brotato'` {
			t.Errorf("Failed to generate hostname query. Output: %v", output)
		}
	}
}

func TestParseHostnameDoesNotMatchRegexCaseSensitive(t *testing.T) {
	toBeTested := []string{
		`Hostname !~ "brotato"`,
		`Hostname!~"brotato"`,
		`hostname !~ "brotato"`,
	}

	for _, testString := range toBeTested {
		output := Parse(testString)
		if output != `hostname !~ 'brotato'` {
			t.Errorf("Failed to generate hostname query. Output: %v", output)
		}
	}
}

func TestParseHostnameMatchRegexCaseInsensitive(t *testing.T) {
	toBeTested := []string{
		`Hostname ~* "brotato"`,
		`Hostname~*"brotato"`,
		`hostname ~* "brotato"`,
	}

	for _, testString := range toBeTested {
		output := Parse(testString)
		if output != `hostname ~* 'brotato'` {
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
		output := Parse(testString)
		if output != `hostname ~ 'brotato'` {
			t.Errorf("Failed to generate hostname query. Output: %v", output)
		}
	}
}

func TestParseJsonTraversal(t *testing.T) {
	// Example JSON:
	// {"/free": {"Swap": {"Free": 0, "Used": 0, "Total": 0}, "Memory": {"Free": 1346609152, "Used": 7243325440, "Total": 8589934592, "ActualFree": 3666075648, "ActualUsed": 4923858944}}}

	toBeTested := `/free.Memory.Free > 10000000`

	output := Parse(toBeTested)
	if output != `data #>> '{/free,Memory.Free}' > '10000000'` {
		t.Errorf("Failed to generate data query. Output: %v", output)
	}
}

func TestParseAnd(t *testing.T) {
	toBeTested := `tags.aaa = bbb AND Hostname~^"brotato" AND /free.Memory.Free > 10000000`

	output := Parse(toBeTested)
	if output != `tags #>> '{aaa}' = 'bbb' and hostname LIKE 'brotato%' and data #>> '{/free,Memory.Free}' > '10000000'` {
		t.Errorf("Failed to generate mixed of tags,hostname, and data query. Output: %v", output)
	}
}
