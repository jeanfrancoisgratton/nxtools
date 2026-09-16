// nxtools
// Unit tests for the pure flag-validation logic backing `repo create`. These
// don't touch the network or cobra: normalizeWritePolicy, isAlpineFormat,
// isAptFormat, checkAlpineSigningRequirement and checkAptSigningRequirement
// are deterministic functions of their string/bool inputs.

package cmd

import "testing"

func TestNormalizeWritePolicy(t *testing.T) {
	cases := map[string]string{
		"allow":      "ALLOW",
		"ALLOW":      "ALLOW",
		"Allow_Once": "ALLOW_ONCE",
		"deny":       "DENY",
		"DENY":       "DENY",
	}
	for in, want := range cases {
		got, err := normalizeWritePolicy(in)
		if err != nil {
			t.Fatalf("normalizeWritePolicy(%q): unexpected error: %v", in, err)
		}
		if got != want {
			t.Errorf("normalizeWritePolicy(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeWritePolicy_Invalid(t *testing.T) {
	if _, err := normalizeWritePolicy("bogus"); err == nil {
		t.Fatal("expected an error for an unsupported write policy")
	}
}

func TestIsAlpineFormat(t *testing.T) {
	for _, f := range []string{"alpine", "Alpine", "ALPINE", "apk", "APK"} {
		if !isAlpineFormat(f) {
			t.Errorf("isAlpineFormat(%q) = false, want true", f)
		}
	}
	for _, f := range []string{"yum", "apt", "docker", ""} {
		if isAlpineFormat(f) {
			t.Errorf("isAlpineFormat(%q) = true, want false", f)
		}
	}
}

func TestIsAptFormat(t *testing.T) {
	for _, f := range []string{"apt", "APT", "Apt"} {
		if !isAptFormat(f) {
			t.Errorf("isAptFormat(%q) = false, want true", f)
		}
	}
	for _, f := range []string{"yum", "alpine", ""} {
		if isAptFormat(f) {
			t.Errorf("isAptFormat(%q) = true, want false", f)
		}
	}
}

func TestCheckAlpineSigningRequirement(t *testing.T) {
	t.Run("sign passed, no keyfile: clean pass", func(t *testing.T) {
		warning, err := checkAlpineSigningRequirement("", true)
		if warning != "" {
			t.Errorf("expected no warning, got %q", warning)
		}
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("sign passed, keyfile also set: warns but doesn't fail", func(t *testing.T) {
		warning, err := checkAlpineSigningRequirement("/path/to/key", true)
		if warning == "" {
			t.Error("expected a warning about --keyfile being ignored")
		}
		if err != nil {
			t.Errorf("expected no hard error, got %v", err)
		}
	})

	t.Run("sign not passed: hard error", func(t *testing.T) {
		_, err := checkAlpineSigningRequirement("", false)
		if err == nil {
			t.Fatal("expected an error when --sign wasn't passed")
		}
	})

	t.Run("sign not passed and keyfile set: both warning and error", func(t *testing.T) {
		warning, err := checkAlpineSigningRequirement("/path/to/key", false)
		if warning == "" {
			t.Error("expected a warning about --keyfile being ignored")
		}
		if err == nil {
			t.Fatal("expected an error when --sign wasn't passed")
		}
	})
}

func TestCheckAptSigningRequirement(t *testing.T) {
	t.Run("sign only: generate mode", func(t *testing.T) {
		mode, err := checkAptSigningRequirement("", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mode != aptSignGenerate {
			t.Errorf("mode = %v, want aptSignGenerate", mode)
		}
	})

	t.Run("keyfile only: keyfile mode", func(t *testing.T) {
		mode, err := checkAptSigningRequirement("/path/to/key", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mode != aptSignKeyfile {
			t.Errorf("mode = %v, want aptSignKeyfile", mode)
		}
	})

	t.Run("both sign and keyfile: rejected as ambiguous", func(t *testing.T) {
		if _, err := checkAptSigningRequirement("/path/to/key", true); err == nil {
			t.Fatal("expected an error when both --sign and -k/--keyfile are given")
		}
	})

	t.Run("neither sign nor keyfile: rejected as missing", func(t *testing.T) {
		if _, err := checkAptSigningRequirement("", false); err == nil {
			t.Fatal("expected an error when neither --sign nor -k/--keyfile is given")
		}
	})
}
