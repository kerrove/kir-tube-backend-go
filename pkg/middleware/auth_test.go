package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go/kir-tube/configs"
	"go/kir-tube/pkg/di"
	"go/kir-tube/pkg/jwt"
)

const testSecret = "test-secret"

// fakeUserProvider implements di.IUserProvider for the middleware tests.
type fakeUserProvider struct {
	user *di.ContextUser
	err  error
}

func (f *fakeUserProvider) FindContextUser(id string) (*di.ContextUser, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.user, nil
}

func testConfig() *configs.Config {
	return &configs.Config{Auth: configs.AuthConfig{Secret: testSecret}}
}

func bearer(t *testing.T, id string) string {
	t.Helper()
	token, err := jwt.NewJWT(testSecret).Create(jwt.JWTData{Id: id}, "1h")
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	return "Bearer " + token
}

func TestIsAuthedRejectsMissingHeader(t *testing.T) {
	provider := &fakeUserProvider{user: &di.ContextUser{ID: "u1"}}
	called := false
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })

	rec := httptest.NewRecorder()
	IsAuthed(next, testConfig(), provider).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if called {
		t.Fatal("next handler ran despite missing auth")
	}
}

func TestIsAuthedRejectsInvalidToken(t *testing.T) {
	provider := &fakeUserProvider{user: &di.ContextUser{ID: "u1"}}
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("next ran") })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer garbage")
	rec := httptest.NewRecorder()
	IsAuthed(next, testConfig(), provider).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestIsAuthedRejectsUnknownUser(t *testing.T) {
	provider := &fakeUserProvider{err: errors.New("not found")}
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("next ran") })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", bearer(t, "missing"))
	rec := httptest.NewRecorder()
	IsAuthed(next, testConfig(), provider).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestIsAuthedPassesUserInContext(t *testing.T) {
	want := &di.ContextUser{ID: "u1", Email: "a@b.c"}
	provider := &fakeUserProvider{user: want}

	var got *di.ContextUser
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got, _ = r.Context().Value(ContextUserKey).(*di.ContextUser)
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", bearer(t, "u1"))
	rec := httptest.NewRecorder()
	IsAuthed(next, testConfig(), provider).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got == nil || got.ID != "u1" {
		t.Fatalf("context user = %+v, want ID u1", got)
	}
}

func TestMaybeAuthedAllowsAnonymous(t *testing.T) {
	provider := &fakeUserProvider{user: &di.ContextUser{ID: "u1"}}
	ran := false
	var got *di.ContextUser
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ran = true
		got, _ = r.Context().Value(ContextUserKey).(*di.ContextUser)
	})

	rec := httptest.NewRecorder()
	MaybeAuthed(next, testConfig(), provider).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	if !ran {
		t.Fatal("next did not run for anonymous request")
	}
	if got != nil {
		t.Fatalf("anonymous request has a context user: %+v", got)
	}
}

func TestMaybeAuthedInjectsUserWhenTokenValid(t *testing.T) {
	provider := &fakeUserProvider{user: &di.ContextUser{ID: "u9"}}
	var got *di.ContextUser
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got, _ = r.Context().Value(ContextUserKey).(*di.ContextUser)
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", bearer(t, "u9"))
	rec := httptest.NewRecorder()
	MaybeAuthed(next, testConfig(), provider).ServeHTTP(rec, req)

	if got == nil || got.ID != "u9" {
		t.Fatalf("context user = %+v, want ID u9", got)
	}
}
