package jwt

import "testing"

func TestCreateAndParseRoundTrip(t *testing.T) {
	j := NewJWT("secret")
	token, err := j.Create(JWTData{Id: "user-1", IsAdmin: true}, "1h")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	valid, data := j.Parse(token)
	if !valid {
		t.Fatal("Parse reported a freshly created token as invalid")
	}
	if data.Id != "user-1" {
		t.Fatalf("Id = %q, want user-1", data.Id)
	}
	if !data.IsAdmin {
		t.Fatal("IsAdmin = false, want true")
	}
}

func TestParseRejectsWrongSecret(t *testing.T) {
	token, err := NewJWT("secret").Create(JWTData{Id: "user-1"}, "1h")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	valid, data := NewJWT("other-secret").Parse(token)
	if valid || data != nil {
		t.Fatal("Parse accepted a token signed with a different secret")
	}
}

func TestParseRejectsExpiredToken(t *testing.T) {
	token, err := NewJWT("secret").Create(JWTData{Id: "user-1"}, "-1h")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if valid, _ := NewJWT("secret").Parse(token); valid {
		t.Fatal("Parse accepted an expired token")
	}
}

func TestParseRejectsGarbage(t *testing.T) {
	if valid, _ := NewJWT("secret").Parse("not.a.jwt"); valid {
		t.Fatal("Parse accepted a malformed token")
	}
}

func TestCreateRejectsBadDuration(t *testing.T) {
	if _, err := NewJWT("secret").Create(JWTData{Id: "x"}, "not-a-duration"); err == nil {
		t.Fatal("Create accepted an invalid duration")
	}
}
