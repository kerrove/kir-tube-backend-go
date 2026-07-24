package auth

import (
	"errors"
	"testing"

	"go/kir-tube/configs"
	"go/kir-tube/internal/user"
	"go/kir-tube/pkg/jwt"
	"go/kir-tube/pkg/password"
)

// mockUserService implements IUserService via injectable func fields.
type mockUserService struct {
	createFn         func(email, pass string) (*user.User, error)
	getByIdFn        func(string) (*user.User, error)
	getByEmailFn     func(string) (*user.User, error)
	getByVerifyFn    func(string) (*user.User, error)
	updateFn         func(*user.User) (*user.User, error)
	lastUpdatedInput *user.User
}

func (m *mockUserService) Create(email, pass string) (*user.User, error) {
	return m.createFn(email, pass)
}
func (m *mockUserService) GetById(id string) (*user.User, error)         { return m.getByIdFn(id) }
func (m *mockUserService) GetByEmail(e string) (*user.User, error)       { return m.getByEmailFn(e) }
func (m *mockUserService) GetByVerifyToken(t string) (*user.User, error) { return m.getByVerifyFn(t) }
func (m *mockUserService) Update(u *user.User) (*user.User, error) {
	m.lastUpdatedInput = u
	return m.updateFn(u)
}

func newService(m *mockUserService) *AuthService {
	return NewAuthService(&AuthServiceDeps{
		UserService: m,
		Config:      &configs.Config{Auth: configs.AuthConfig{Secret: "test-secret"}},
	})
}

func userWithPassword(t *testing.T, id, email, plain string) *user.User {
	t.Helper()
	hash, err := password.Encode(plain)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	u := &user.User{Email: email, Password: hash}
	u.ID = id
	return u
}

func TestLoginSuccess(t *testing.T) {
	stored := userWithPassword(t, "u1", "a@b.c", "secret")
	svc := newService(&mockUserService{
		getByEmailFn: func(string) (*user.User, error) { return stored, nil },
	})

	res, err := svc.Login(AuthRequest{Email: "a@b.c", Password: "secret"})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if res.AccessToken == "" || res.RefreshToken == "" {
		t.Fatal("Login did not issue tokens")
	}
	valid, data := jwt.NewJWT("test-secret").Parse(res.AccessToken)
	if !valid || data.Id != "u1" {
		t.Fatalf("access token invalid or wrong subject: %+v", data)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	stored := userWithPassword(t, "u1", "a@b.c", "secret")
	svc := newService(&mockUserService{
		getByEmailFn: func(string) (*user.User, error) { return stored, nil },
	})

	_, err := svc.Login(AuthRequest{Email: "a@b.c", Password: "wrong"})
	if err == nil || err.Error() != ErrWrongCredential {
		t.Fatalf("err = %v, want %q", err, ErrWrongCredential)
	}
}

func TestLoginUnknownUser(t *testing.T) {
	svc := newService(&mockUserService{
		getByEmailFn: func(string) (*user.User, error) { return nil, errors.New("not found") },
	})
	if _, err := svc.Login(AuthRequest{Email: "x@y.z", Password: "p"}); err == nil {
		t.Fatal("Login accepted an unknown user")
	}
}

func TestRegisterNewUser(t *testing.T) {
	created := userWithPassword(t, "new-id", "new@b.c", "p")
	svc := newService(&mockUserService{
		getByEmailFn: func(string) (*user.User, error) { return nil, errors.New("not found") },
		createFn:     func(string, string) (*user.User, error) { return created, nil },
	})

	res, err := svc.Register(AuthRequest{Email: "new@b.c", Password: "p"})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if res.User.ID != "new-id" || res.AccessToken == "" {
		t.Fatalf("unexpected register result: %+v", res)
	}
}

func TestRegisterExistingUser(t *testing.T) {
	stored := userWithPassword(t, "u1", "a@b.c", "secret")
	svc := newService(&mockUserService{
		getByEmailFn: func(string) (*user.User, error) { return stored, nil },
	})

	_, err := svc.Register(AuthRequest{Email: "a@b.c", Password: "secret"})
	if err == nil || err.Error() != ErrUserExist {
		t.Fatalf("err = %v, want %q", err, ErrUserExist)
	}
}

func TestGetNewTokensValid(t *testing.T) {
	stored := userWithPassword(t, "u1", "a@b.c", "secret")
	svc := newService(&mockUserService{
		getByIdFn: func(string) (*user.User, error) { return stored, nil },
	})

	refresh, err := jwt.NewJWT("test-secret").Create(jwt.JWTData{Id: "u1"}, "1h")
	if err != nil {
		t.Fatalf("create refresh: %v", err)
	}

	res, err := svc.GetNewTokens(refresh)
	if err != nil {
		t.Fatalf("GetNewTokens: %v", err)
	}
	if res.AccessToken == "" || res.User.ID != "u1" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestGetNewTokensInvalid(t *testing.T) {
	svc := newService(&mockUserService{})
	if _, err := svc.GetNewTokens("garbage"); err == nil {
		t.Fatal("GetNewTokens accepted an invalid refresh token")
	}
}

func TestVerifyEmailClearsToken(t *testing.T) {
	token := "verify-token"
	tok := token
	stored := &user.User{Email: "a@b.c", VerificationToken: &tok}
	stored.ID = "u1"

	m := &mockUserService{
		getByVerifyFn: func(string) (*user.User, error) { return stored, nil },
		updateFn:      func(u *user.User) (*user.User, error) { return u, nil },
	}
	svc := newService(m)

	msg, err := svc.verifyEmail(token)
	if err != nil {
		t.Fatalf("verifyEmail: %v", err)
	}
	if msg != "Email verified!" {
		t.Fatalf("msg = %q", msg)
	}
	if m.lastUpdatedInput == nil || m.lastUpdatedInput.VerificationToken != nil {
		t.Fatal("verifyEmail did not clear the verification token before update")
	}
}
