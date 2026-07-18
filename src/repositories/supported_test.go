// nxtools
// Unit tests for supported-format helpers.

package repositories

import (
	"strings"
	"testing"
)

func TestGetLogicalVal(t *testing.T) {
	if got := getLogicalVal(&TrueVal); !strings.Contains(got, "yes") {
		t.Errorf("getLogicalVal(&TrueVal) = %q, want to contain 'yes'", got)
	}
	if got := getLogicalVal(&FalseVal); !strings.Contains(got, "no") {
		t.Errorf("getLogicalVal(&FalseVal) = %q, want to contain 'no'", got)
	}
	if got := getLogicalVal(nil); !strings.Contains(got, "n/a") {
		t.Errorf("getLogicalVal(nil) = %q, want to contain 'n/a'", got)
	}

	// A pointer to some other bool is neither TrueVal nor FalseVal -> n/a.
	other := true
	if got := getLogicalVal(&other); !strings.Contains(got, "n/a") {
		t.Errorf("getLogicalVal(&other) = %q, want to contain 'n/a'", got)
	}
}

func TestSupportedFormatsIncludesAlpine(t *testing.T) {
	var alpine *FormatStatusStruct
	for i := range supportedformats {
		if supportedformats[i].Name == "alpine" {
			alpine = &supportedformats[i]
			break
		}
	}
	if alpine == nil {
		t.Fatal("alpine is missing from supportedformats")
	}
	if alpine.SupportsHosted == nil || *alpine.SupportsHosted != true {
		t.Error("alpine should advertise hosted support")
	}
}
