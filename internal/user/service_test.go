package user

import (
	"errors"
	"testing"

	"go/kir-tube/pkg/di"
	"go/kir-tube/pkg/password"
)

type mockUserRepo struct {
	created    *User
	updated    *User
	createFn   func(*User) (*User, error)
	updateFn   func(*User) (*User, error)
	findByIdFn func(string) (*User, error)
}

func (m *mockUserRepo) Create(u *User) (*User, error) {
	m.created = u
	if m.createFn != nil {
		return m.createFn(u)
	}
	return u, nil
}
func (m *mockUserRepo) Update(u *User) (*User, error) {
	m.updated = u
	if m.updateFn != nil {
		return m.updateFn(u)
	}
	return u, nil
}
func (m *mockUserRepo) FindById(id string) (*User, error) {
	if m.findByIdFn != nil {
		return m.findByIdFn(id)
	}
	return nil, errors.New("not stubbed")
}
func (m *mockUserRepo) FindByEmail(string) (*User, error)       { return nil, errors.New("nope") }
func (m *mockUserRepo) FindByVerifyToken(string) (*User, error) { return nil, errors.New("nope") }
func (m *mockUserRepo) FindAll() []User                         { return nil }

type mockVideoRepo struct {
	subscribed    []di.SubscribedVideo
	liked         []di.SubscribedVideo
	subscriptions []di.SubscriptionChannel
}

func (m *mockVideoRepo) FindSubscribedVideos(string) ([]di.SubscribedVideo, error) {
	return m.subscribed, nil
}
func (m *mockVideoRepo) FindLikedVideos(string) ([]di.SubscribedVideo, error) { return m.liked, nil }
func (m *mockVideoRepo) FindSubscriptions(string) ([]di.SubscriptionChannel, error) {
	return m.subscriptions, nil
}

func newUserService(repo UserRepositoryPort, vids di.IVideoRepository) *UserService {
	return NewUserService(&UserServiceDeps{UserRepository: repo, VideoRepository: vids})
}

func TestCreateHashesPassword(t *testing.T) {
	repo := &mockUserRepo{}
	svc := newUserService(repo, &mockVideoRepo{})

	if _, err := svc.Create("a@b.c", "plaintext"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if repo.created == nil {
		t.Fatal("repository Create was not called")
	}
	if repo.created.Password == "plaintext" {
		t.Fatal("password stored in plaintext")
	}
	if !password.Validate(repo.created.Password, "plaintext") {
		t.Fatal("stored hash does not validate against the original password")
	}
	if repo.created.Email != "a@b.c" {
		t.Fatalf("email = %q", repo.created.Email)
	}
}

func TestGetProfileAggregates(t *testing.T) {
	stored := &User{Email: "a@b.c"}
	stored.ID = "u1"
	repo := &mockUserRepo{findByIdFn: func(string) (*User, error) { return stored, nil }}
	vids := &mockVideoRepo{
		subscribed:    []di.SubscribedVideo{{ID: "v1"}},
		liked:         []di.SubscribedVideo{{ID: "v2"}},
		subscriptions: []di.SubscriptionChannel{{ID: "c1"}},
	}
	svc := newUserService(repo, vids)

	profile, err := svc.GetProfile("u1")
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if profile.User.ID != "u1" {
		t.Fatalf("user id = %q", profile.User.ID)
	}
	if len(profile.SubscribedVideos) != 1 || len(profile.Likes) != 1 || len(profile.Subscriptions) != 1 {
		t.Fatalf("aggregates not populated: %+v", profile)
	}
}

func TestGetProfileUserNotFound(t *testing.T) {
	repo := &mockUserRepo{findByIdFn: func(string) (*User, error) { return nil, errors.New("nope") }}
	svc := newUserService(repo, &mockVideoRepo{})
	if _, err := svc.GetProfile("missing"); err == nil {
		t.Fatal("GetProfile returned no error for a missing user")
	}
}

func TestUpdateProfileChangesNameAndPassword(t *testing.T) {
	stored := &User{Email: "a@b.c", Password: "old-hash"}
	stored.ID = "u1"
	repo := &mockUserRepo{findByIdFn: func(string) (*User, error) { return stored, nil }}
	svc := newUserService(repo, &mockVideoRepo{})

	newName := "Kir"
	newPass := "new-password"
	_, err := svc.UpdateProfile("u1", &UpdateProfileReq{Name: &newName, Password: &newPass})
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	if repo.updated == nil {
		t.Fatal("repository Update was not called")
	}
	if repo.updated.Name == nil || *repo.updated.Name != "Kir" {
		t.Fatalf("name not updated: %+v", repo.updated.Name)
	}
	if repo.updated.Password == "old-hash" || !password.Validate(repo.updated.Password, "new-password") {
		t.Fatal("password was not re-hashed on update")
	}
}
