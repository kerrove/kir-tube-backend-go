package playlist

import (
	"errors"
	"testing"

	"gorm.io/gorm"
)

type mockPlaylistRepo struct {
	byUser     *[]Playlist
	byId       *Playlist
	findByUser func(string) (*[]Playlist, error)
	findById   func(string) (*Playlist, error)
	toggleFn   func(playlistId, videoId, userId string) (*ToggleVideoResponse, error)
	createFn   func(userId string, body *PlaylistRequest) (*Playlist, error)
}

func (m *mockPlaylistRepo) FindByUserId(id string) (*[]Playlist, error) {
	if m.findByUser != nil {
		return m.findByUser(id)
	}
	return m.byUser, nil
}
func (m *mockPlaylistRepo) FindById(id string) (*Playlist, error) {
	if m.findById != nil {
		return m.findById(id)
	}
	return m.byId, nil
}
func (m *mockPlaylistRepo) ToggleVideo(p, v, u string) (*ToggleVideoResponse, error) {
	return m.toggleFn(p, v, u)
}
func (m *mockPlaylistRepo) Create(userId string, body *PlaylistRequest) (*Playlist, error) {
	return m.createFn(userId, body)
}

func newPlaylistService(repo PlaylistRepositoryPort) *PlaylistService {
	return NewPlaylistService(&PlaylistServiceDeps{PlaylistRepository: repo})
}

func TestGetPlaylistByIdNotFoundMapping(t *testing.T) {
	svc := newPlaylistService(&mockPlaylistRepo{
		findById: func(string) (*Playlist, error) { return nil, gorm.ErrRecordNotFound },
	})
	_, err := svc.GetPlaylistById("missing")
	if err == nil || err.Error() != ErrPlaylistNotExist {
		t.Fatalf("err = %v, want %q", err, ErrPlaylistNotExist)
	}
}

func TestGetUserPlaylistNotFoundMapping(t *testing.T) {
	svc := newPlaylistService(&mockPlaylistRepo{
		findByUser: func(string) (*[]Playlist, error) { return nil, gorm.ErrRecordNotFound },
	})
	if _, err := svc.GetUserPlaylist("u1"); err == nil || err.Error() != ErrPlaylistNotExist {
		t.Fatalf("err = %v, want %q", err, ErrPlaylistNotExist)
	}
}

func TestCreateDelegates(t *testing.T) {
	want := &Playlist{Title: "Favs"}
	var gotUser string
	svc := newPlaylistService(&mockPlaylistRepo{
		createFn: func(userId string, body *PlaylistRequest) (*Playlist, error) {
			gotUser = userId
			return want, nil
		},
	})
	got, err := svc.Create("u1", &PlaylistRequest{Title: "Favs"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got != want || gotUser != "u1" {
		t.Fatalf("Create delegated incorrectly: user=%q got=%+v", gotUser, got)
	}
}

func TestToggleVideoDelegates(t *testing.T) {
	want := &ToggleVideoResponse{}
	svc := newPlaylistService(&mockPlaylistRepo{
		toggleFn: func(p, v, u string) (*ToggleVideoResponse, error) {
			if p != "p1" || v != "v1" || u != "u1" {
				return nil, errors.New("wrong args")
			}
			return want, nil
		},
	})
	got, err := svc.ToggleVideo("u1", "p1", "v1")
	if err != nil {
		t.Fatalf("ToggleVideo: %v", err)
	}
	if got != want {
		t.Fatal("ToggleVideo did not return the repository result")
	}
}
