// nxtools
// Unit tests for filename-based asset fetching (raw / Arch Linux packages).

package assets

import "testing"

func TestFilenameMatchesPackage(t *testing.T) {
	cases := []struct {
		basename string
		pkg      string
		want     bool
	}{
		{"stubber-2.6.0-2-x86_64.pkg.tar.zst", "stubber", true},
		{"stubber_2.6.0-1_amd64.deb", "stubber", true},
		{"stubber-2.6.0-3.x86_64.rpm", "stubber", true},
		{"stubber", "stubber", true},
		{"Stubber-2.6.0-2-x86_64.pkg.tar.zst", "stubber", true}, // case-insensitive
		{"stubber-utils-1.0-1-x86_64.pkg.tar.zst", "stubber", false},
		{"stubberd-1.0-1-x86_64.pkg.tar.zst", "stubber", false},
		{"otherpkg-1.0.pkg.tar.zst", "stubber", false},
		{"stubber-x86_64.pkg.tar.zst", "stubber", false}, // no version digit after separator
		{"", "stubber", false},
		{"stubber-2.6.0.pkg.tar.zst", "", false},
	}
	for _, c := range cases {
		if got := filenameMatchesPackage(c.basename, c.pkg); got != c.want {
			t.Errorf("filenameMatchesPackage(%q, %q) = %v, want %v", c.basename, c.pkg, got, c.want)
		}
	}
}

func TestNaturalCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"stubber-2.9.0-1-x86_64.pkg.tar.zst", "stubber-2.10.0-1-x86_64.pkg.tar.zst", -1},
		{"stubber-2.6.0-2-x86_64.pkg.tar.zst", "stubber-2.6.0-1-x86_64.pkg.tar.zst", 1},
		{"stubber-2.6.0-2", "stubber-2.6.0-2", 0},
		{"v1", "v10", -1},
		{"a", "b", -1},
		{"stubber-2.6.0.pkg.tar.zst", "stubber-2.6.0.pkg.tar.zst.sig", -1},
	}
	for _, c := range cases {
		if got := naturalCompare(c.a, c.b); got != c.want {
			t.Errorf("naturalCompare(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
