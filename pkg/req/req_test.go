package request

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type sample struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required"`
}

func body(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}

func TestDecodeValidJSON(t *testing.T) {
	got, err := Decode[sample](body(`{"email":"a@b.c","name":"kir"}`))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got.Email != "a@b.c" || got.Name != "kir" {
		t.Fatalf("decoded = %+v", got)
	}
}

func TestDecodeInvalidJSON(t *testing.T) {
	if _, err := Decode[sample](body(`{not json`)); err == nil {
		t.Fatal("Decode accepted malformed JSON")
	}
}

func TestIsValid(t *testing.T) {
	if err := IsValid(sample{Email: "a@b.c", Name: "kir"}); err != nil {
		t.Fatalf("IsValid rejected a valid struct: %v", err)
	}
	if err := IsValid(sample{Email: "not-an-email", Name: "kir"}); err == nil {
		t.Fatal("IsValid accepted a bad email")
	}
	if err := IsValid(sample{Email: "a@b.c"}); err == nil {
		t.Fatal("IsValid accepted a missing required field")
	}
}

func TestHandleBodyValid(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"email":"a@b.c","name":"kir"}`))
	w := httptest.NewRecorder()
	var rw http.ResponseWriter = w

	got, err := HandleBody[sample](&rw, r)
	if err != nil {
		t.Fatalf("HandleBody: %v", err)
	}
	if got.Email != "a@b.c" {
		t.Fatalf("got = %+v", got)
	}
}

func TestHandleBodyRejectsInvalidBody(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"email":"bad","name":""}`))
	w := httptest.NewRecorder()
	var rw http.ResponseWriter = w

	if _, err := HandleBody[sample](&rw, r); err == nil {
		t.Fatal("HandleBody accepted an invalid payload")
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestHandleBodyRejectsBrokenJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{broken`))
	w := httptest.NewRecorder()
	var rw http.ResponseWriter = w

	if _, err := HandleBody[sample](&rw, r); err == nil {
		t.Fatal("HandleBody accepted broken JSON")
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}
