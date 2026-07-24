package configs

import "testing"

func TestParseBool(t *testing.T) {
	cases := map[string]bool{
		"true":  true,
		"1":     true,
		"false": false,
		"0":     false,
		"":      false,
		"nope":  false,
	}
	for in, want := range cases {
		if got := parseBool(in); got != want {
			t.Errorf("parseBool(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestGetEnvDefault(t *testing.T) {
	t.Setenv("SOME_KEY", "value")
	if got := getEnvDefault("SOME_KEY", "fallback"); got != "value" {
		t.Fatalf("got %q, want value", got)
	}
	if got := getEnvDefault("MISSING_KEY_XYZ", "fallback"); got != "fallback" {
		t.Fatalf("got %q, want fallback", got)
	}
}

func TestLoadConfigMapsEnv(t *testing.T) {
	t.Setenv("DSN", "test-dsn")
	t.Setenv("JWT_SECRET", "s3cr3t")
	t.Setenv("SECURE_COOKIES", "true")
	t.Setenv("PORT", "9999")
	t.Setenv("DOMAIN", "example.com")
	t.Setenv("CLIENT_URL", "http://client")
	t.Setenv("MINIO_ENDPOINT", "minio:9000")
	t.Setenv("MINIO_ROOT_USER", "minio")
	t.Setenv("MINIO_ROOT_PASSWORD", "secret")
	t.Setenv("MINIO_USE_SSL", "true")

	c := LoadConfig()

	if c.Db.Dsn != "test-dsn" {
		t.Errorf("Dsn = %q", c.Db.Dsn)
	}
	if c.Auth.Secret != "s3cr3t" || !c.Auth.SecureCookies {
		t.Errorf("Auth = %+v", c.Auth)
	}
	if c.Network.Port != "9999" || c.Network.Domain != "example.com" || c.Network.ClientUrl != "http://client" {
		t.Errorf("Network = %+v", c.Network)
	}
	if c.Storage.Endpoint != "minio:9000" || c.Storage.AccessKey != "minio" ||
		c.Storage.SecretKey != "secret" || !c.Storage.UseSSL {
		t.Errorf("Storage = %+v", c.Storage)
	}
}

func TestLoadConfigBucketDefault(t *testing.T) {
	t.Setenv("MINIO_BUCKET", "")
	if c := LoadConfig(); c.Storage.Bucket != "kir-tube" {
		t.Fatalf("Bucket = %q, want kir-tube", c.Storage.Bucket)
	}
}
