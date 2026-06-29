package validate

import (
	"testing"
)

func TestValidateCommand(t *testing.T) {
	cases := []struct {
		cmd  string
		want bool
	}{
		{"cr", true},
		{"cache:rebuild", true},
		{"cc render", true},
		{"cc plugin", true},
		{"php -l", true},
		{"yaml", true},
		{"unknown", false},
		{"rm -rf", false},
	}
	for _, tc := range cases {
		got := validateCommand(tc.cmd)
		if got != tc.want {
			t.Errorf("validateCommand(%q) = %v, want %v", tc.cmd, got, tc.want)
		}
	}
}

func TestFindPHPCS(t *testing.T) {
	got := findPHPCS(t.TempDir())
	if got != "" && got != "phpcs" {
		t.Logf("phpcs found at: %s", got)
	}
}

func TestValidate_NoConfig(t *testing.T) {
	// This will work with an empty dir
	result := Validate(t.TempDir())
	if result.Pass {
		t.Error("expected validation to fail in empty directory")
	}
	if len(result.Entries) == 0 {
		t.Error("expected at least some entries")
	}
}
