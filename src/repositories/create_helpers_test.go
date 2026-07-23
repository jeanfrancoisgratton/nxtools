// nxtools
// Unit tests for the per-format payload builders.
//
// These builders read package-level configuration globals, so each test snapshots
// and restores the globals it touches to avoid cross-test contamination.

package repositories

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

// genPKCS1RSAKeyPEM generates a throwaway RSA key and returns it PEM-encoded
// in PKCS1 ("RSA PRIVATE KEY") form, alongside the key itself for comparison.
func genPKCS1RSAKeyPEM(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pemStr := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}))
	return priv, pemStr
}

// genPKCS8RSAKeyPEM generates a throwaway RSA key and returns it PEM-encoded
// in PKCS8 ("PRIVATE KEY") form, alongside the key itself for comparison.
func genPKCS8RSAKeyPEM(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	pemStr := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	return priv, pemStr
}

// writeKeyFile writes a throwaway signing key and points RepoSigningFile at it.
func writeKeyFile(t *testing.T, contents string) {
	t.Helper()
	fp := filepath.Join(t.TempDir(), "key.pem")
	if err := os.WriteFile(fp, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	prev := RepoSigningFile
	RepoSigningFile = fp
	t.Cleanup(func() { RepoSigningFile = prev })
}

func TestReadFileAsString(t *testing.T) {
	fp := filepath.Join(t.TempDir(), "f.txt")
	if err := os.WriteFile(fp, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := readFileAsString(fp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello world" {
		t.Fatalf("readFileAsString = %q", got)
	}

	if _, err := readFileAsString(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestCreateApt(t *testing.T) {
	writeKeyFile(t, "APT-KEY-DATA")
	swap(t, &RepoSigningPassphrase, "pp")
	swap(t, &RepoAptDistro, "bookworm")
	swap(t, &StorageWritePolicy, "ALLOW")
	swapBool(t, &StorageStrictContentValidation, true)

	payload, err := createApt("aptrepo", "aptblob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out AptRepoSettingsStruct
	if e := json.Unmarshal(payload, &out); e != nil {
		t.Fatalf("payload not valid JSON: %v", e)
	}
	if out.Name != "aptrepo" || !out.Online {
		t.Errorf("name/online = %q / %v", out.Name, out.Online)
	}
	if out.Storage.BlobStoreName != "aptblob" || out.Storage.WritePolicy != "ALLOW" {
		t.Errorf("storage = %+v", out.Storage)
	}
	if out.Apt.Distribution != "bookworm" {
		t.Errorf("distribution = %q", out.Apt.Distribution)
	}
	if out.AptSigning.Keypair != "APT-KEY-DATA" || out.AptSigning.Passphrase != "pp" {
		t.Errorf("aptSigning = %+v", out.AptSigning)
	}
}

func TestCreateApt_MissingKeyFile(t *testing.T) {
	prev := RepoSigningFile
	RepoSigningFile = filepath.Join(t.TempDir(), "nope.pem")
	t.Cleanup(func() { RepoSigningFile = prev })

	if _, err := createApt("r", "b"); err == nil {
		t.Fatal("expected error when signing key file is missing")
	}
}

func TestCreateAlpine(t *testing.T) {
	_, pkcs1Pem := genPKCS1RSAKeyPEM(t)
	writeKeyFile(t, pkcs1Pem)
	swap(t, &RepoSigningPassphrase, "secret")
	swap(t, &StorageWritePolicy, "ALLOW_ONCE")
	swapBool(t, &StorageStrictContentValidation, true)

	payload, err := createAlpine("alpinerepo", "alpineblob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out AlpineRepoSettingsStruct
	if e := json.Unmarshal(payload, &out); e != nil {
		t.Fatalf("payload not valid JSON: %v", e)
	}
	if out.Name != "alpinerepo" || !out.Online {
		t.Errorf("name/online = %q / %v", out.Name, out.Online)
	}
	if out.Storage.BlobStoreName != "alpineblob" || out.Storage.WritePolicy != "ALLOW_ONCE" {
		t.Errorf("storage = %+v", out.Storage)
	}
	// A key already in PKCS1 form must pass through unchanged.
	if out.AlpineSigning.Keypair != pkcs1Pem || out.AlpineSigning.Passphrase != "secret" {
		t.Errorf("alpineSigning = %+v", out.AlpineSigning)
	}

	// The Alpine payload must carry alpineSigning and must NOT leak apt-only fields.
	var keys map[string]json.RawMessage
	if e := json.Unmarshal(payload, &keys); e != nil {
		t.Fatal(e)
	}
	if _, ok := keys["alpineSigning"]; !ok {
		t.Error("payload missing alpineSigning block")
	}
	if _, ok := keys["apt"]; ok {
		t.Error("alpine payload must not contain an apt block")
	}
	if _, ok := keys["aptSigning"]; ok {
		t.Error("alpine payload must not contain an aptSigning block")
	}
}

func TestCreateAlpine_ConvertsPKCS8ToPKCS1(t *testing.T) {
	priv, pkcs8Pem := genPKCS8RSAKeyPEM(t)
	writeKeyFile(t, pkcs8Pem)

	payload, err := createAlpine("r", "b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out AlpineRepoSettingsStruct
	if e := json.Unmarshal(payload, &out); e != nil {
		t.Fatalf("payload not valid JSON: %v", e)
	}

	block, _ := pem.Decode([]byte(out.AlpineSigning.Keypair))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		t.Fatalf("expected key converted to PKCS1 PEM, got block = %+v", block)
	}
	parsed, e := x509.ParsePKCS1PrivateKey(block.Bytes)
	if e != nil {
		t.Fatalf("converted key does not parse as PKCS1: %v", e)
	}
	if parsed.N.Cmp(priv.N) != 0 {
		t.Error("converted key material does not match original key")
	}
}

func TestCreateAlpine_RejectsNonRSAKey(t *testing.T) {
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(ecKey)
	if err != nil {
		t.Fatal(err)
	}
	ecPem := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	writeKeyFile(t, ecPem)

	if _, err := createAlpine("r", "b"); err == nil {
		t.Fatal("expected error for non-RSA signing key")
	}
}

func TestCreateAlpine_RejectsInvalidPEM(t *testing.T) {
	writeKeyFile(t, "this is not a PEM-encoded key")

	if _, err := createAlpine("r", "b"); err == nil {
		t.Fatal("expected error for non-PEM signing key file")
	}
}

func TestCreateAlpine_MissingKeyFile(t *testing.T) {
	prev := RepoSigningFile
	RepoSigningFile = filepath.Join(t.TempDir(), "nope.pem")
	t.Cleanup(func() { RepoSigningFile = prev })

	if _, err := createAlpine("r", "b"); err == nil {
		t.Fatal("expected error when signing key file is missing")
	}
}

func TestCreateYum(t *testing.T) {
	swapUint(t, &YumRepodataDepth, 3)

	for _, policy := range []string{"PERMISSIVE", "STRICT", "permissive", "strict"} {
		swap(t, &YumDeployPolicy, policy)
		payload, err := createYum("yumrepo", "yumblob")
		if err != nil {
			t.Fatalf("policy %q: unexpected error: %v", policy, err)
		}
		var out YumRepoSettingsStruct
		if e := json.Unmarshal(payload, &out); e != nil {
			t.Fatalf("policy %q: bad JSON: %v", policy, e)
		}
		if out.Yum.DeployPolicy != policy || out.Yum.RepodataDepth != 3 {
			t.Errorf("policy %q: yum = %+v", policy, out.Yum)
		}
	}
}

func TestCreateYum_InvalidPolicy(t *testing.T) {
	swap(t, &YumDeployPolicy, "bogus")
	if _, err := createYum("r", "b"); err == nil {
		t.Fatal("expected error for unsupported deploy policy")
	}
}

func TestCreateMaven(t *testing.T) {
	swap(t, &MavenVersionPolicy, "RELEASE")
	swap(t, &MavenLayoutPolicy, "STRICT")
	swap(t, &RepoContentDisposition, "INLINE")

	payload, err := createMaven("mvnrepo", "mvnblob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out MavenRepoSettingsStruct
	if e := json.Unmarshal(payload, &out); e != nil {
		t.Fatalf("bad JSON: %v", e)
	}
	if out.Maven.VersionPolicy != "RELEASE" || out.Maven.LayoutPolicy != "STRICT" || out.Maven.ContentDisposition != "INLINE" {
		t.Errorf("maven = %+v", out.Maven)
	}
}

func TestCreateMaven_InvalidInputs(t *testing.T) {
	swap(t, &MavenVersionPolicy, "RELEASE")
	swap(t, &MavenLayoutPolicy, "STRICT")
	swap(t, &RepoContentDisposition, "INLINE")

	swap(t, &MavenVersionPolicy, "bogus")
	if _, err := createMaven("r", "b"); err == nil {
		t.Error("expected error for bad version policy")
	}
	swap(t, &MavenVersionPolicy, "RELEASE")

	swap(t, &MavenLayoutPolicy, "bogus")
	if _, err := createMaven("r", "b"); err == nil {
		t.Error("expected error for bad layout policy")
	}
	swap(t, &MavenLayoutPolicy, "STRICT")

	swap(t, &RepoContentDisposition, "bogus")
	if _, err := createMaven("r", "b"); err == nil {
		t.Error("expected error for bad content disposition")
	}
}

func TestCreateDocker(t *testing.T) {
	swapUint(t, &DockerHttpPort, 8082)
	swapUint(t, &DockerHttpsPort, 0)
	swap(t, &DockerSubdomain, "")
	swapBool(t, &DockerV1Enabled, true)
	swapBool(t, &DockerForceBasicAuth, true)
	swapBool(t, &DockerPathEnabled, true)

	payload, err := createDocker("dockerrepo", "dockerblob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out DockerRepoSettingsStruct
	if e := json.Unmarshal(payload, &out); e != nil {
		t.Fatalf("bad JSON: %v", e)
	}
	if out.Docker.HttpPort != 8082 || !out.Docker.V1Enabled || !out.Docker.ForceBasicAuth || !out.Docker.PathEnabled {
		t.Errorf("docker = %+v", out.Docker)
	}
}

func TestCreateDocker_NoConnectionConfig(t *testing.T) {
	swapUint(t, &DockerHttpPort, 0)
	swapUint(t, &DockerHttpsPort, 0)
	swap(t, &DockerSubdomain, "")

	if _, err := createDocker("r", "b"); err == nil {
		t.Fatal("expected error when no http/https port or subdomain is set")
	}
}

func TestCreateGeneric(t *testing.T) {
	swap(t, &StorageWritePolicy, "ALLOW")

	// Non-npm format keeps the configured strict-validation flag (false here).
	swap(t, &RepoFormat, "raw")
	swapBool(t, &StorageStrictContentValidation, false)
	payload, err := createGeneric("rawrepo", "rawblob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out HostedRepoCommonAttributesStruct
	if e := json.Unmarshal(payload, &out); e != nil {
		t.Fatalf("bad JSON: %v", e)
	}
	if out.Storage.BlobStoreName != "rawblob" {
		t.Errorf("blob = %q", out.Storage.BlobStoreName)
	}
	if out.Storage.StrictContentTypeValidation {
		t.Error("raw format should keep strict validation false")
	}

	// npm forces strict validation on regardless of the configured flag.
	swap(t, &RepoFormat, "npm")
	swapBool(t, &StorageStrictContentValidation, false)
	payload, err = createGeneric("npmrepo", "npmblob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e := json.Unmarshal(payload, &out); e != nil {
		t.Fatalf("bad JSON: %v", e)
	}
	if !out.Storage.StrictContentTypeValidation {
		t.Error("npm format must force strict validation true")
	}
}

// --- small typed swap helpers that restore globals after the test ---

func swap(t *testing.T, p *string, v string) {
	t.Helper()
	prev := *p
	*p = v
	t.Cleanup(func() { *p = prev })
}

func swapBool(t *testing.T, p *bool, v bool) {
	t.Helper()
	prev := *p
	*p = v
	t.Cleanup(func() { *p = prev })
}

func swapUint(t *testing.T, p *uint, v uint) {
	t.Helper()
	prev := *p
	*p = v
	t.Cleanup(func() { *p = prev })
}
