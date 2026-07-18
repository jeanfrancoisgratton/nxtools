// nxtools
// Unit tests for asset listing/filtering helpers.

package assets

import (
	"testing"

	"nxtools/shared"
)

func TestShouldIncludeAsset(t *testing.T) {
	cases := []struct {
		format string
		path   string
		want   bool
	}{
		{"apt", "/pool/main/foo.deb", true},
		{"apt", "/pool/main/foo.udeb", true},
		{"apt", "/dists/stable/Release", false},
		{"alpine", "/x86_64/foo-1.0.apk", true},
		{"alpine", "/x86_64/APKINDEX.tar.gz", false},
		{"yum", "/Packages/foo.rpm", true},
		{"yum", "/repodata/repomd.xml", false},
		{"helm", "/charts/foo-1.0.tgz", true},
		{"helm", "/index.yaml", false},
		{"docker", "/v2/manifests/sha256", true},
		{"raw", "/anything/at/all", true},
		{"somethingelse", "/whatever", true},
	}
	for _, c := range cases {
		item := shared.AssetSummary{Path: c.path}
		if got := shouldIncludeAsset(c.format, item); got != c.want {
			t.Errorf("shouldIncludeAsset(%q, %q) = %v, want %v", c.format, c.path, got, c.want)
		}
	}
}

func TestFilterAssetsByFormat(t *testing.T) {
	items := []shared.AssetSummary{
		{Path: "/pool/foo.deb"},
		{Path: "/dists/Release"},
		{Path: "/pool/bar.deb"},
	}
	got := filterAssetsByFormat("apt", items)
	if len(got) != 2 {
		t.Fatalf("filtered %d assets, want 2", len(got))
	}
}

func TestDisplayAssetName(t *testing.T) {
	cases := map[string]string{
		"/a/b/c.rpm": "c.rpm",
		"foo.apk":    "foo.apk",
		"":           "",
		"   ":        "",
	}
	for in, want := range cases {
		if got := displayAssetName(shared.AssetSummary{Path: in}); got != want {
			t.Errorf("displayAssetName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildComponentIdentityKey(t *testing.T) {
	cases := []struct {
		name      string
		component ComponentSummary
		want      string
	}{
		{
			name:      "group and name",
			component: ComponentSummary{Format: "npm", Group: "@scope", Name: "pkg"},
			want:      "npm|@scope|pkg",
		},
		{
			name:      "name only",
			component: ComponentSummary{Format: "pypi", Name: "pkg"},
			want:      "pypi||pkg",
		},
		{
			name: "path derived",
			component: ComponentSummary{
				Format:  "raw",
				Version: "1.0",
				Assets:  []shared.AssetSummary{{Path: "/a/1.0/foo.txt"}},
			},
			want: "raw|/a/foo.txt",
		},
		{
			name:      "fallback to id",
			component: ComponentSummary{Format: "x", ID: "abc123"},
			want:      "x|abc123",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := buildComponentIdentityKey(c.component); got != c.want {
				t.Fatalf("buildComponentIdentityKey = %q, want %q", got, c.want)
			}
		})
	}
}
