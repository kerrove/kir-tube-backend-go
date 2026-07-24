package password

import "testing"

func TestEncodeProducesVerifiableHash(t *testing.T) {
	hash, err := Encode("qwerty123")
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("Encode returned empty hash")
	}
	if hash == "qwerty123" {
		t.Fatal("Encode returned the plaintext password")
	}
	if !Validate(hash, "qwerty123") {
		t.Fatal("Validate rejected the correct password")
	}
}

func TestValidateRejectsWrongPassword(t *testing.T) {
	hash, err := Encode("correct-horse")
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}
	if Validate(hash, "wrong-password") {
		t.Fatal("Validate accepted an incorrect password")
	}
}

func TestValidateRejectsGarbageHash(t *testing.T) {
	if Validate("not-a-bcrypt-hash", "whatever") {
		t.Fatal("Validate accepted a malformed hash")
	}
}

func TestEncodeIsSalted(t *testing.T) {
	a, _ := Encode("same")
	b, _ := Encode("same")
	if a == b {
		t.Fatal("two hashes of the same password are identical; salt missing")
	}
}
