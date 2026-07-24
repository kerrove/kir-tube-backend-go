package channel

import (
	"errors"
	"testing"

	"gorm.io/gorm"
)

type mockChannelRepo struct {
	all         *[]Channel
	details     *ChannelDetails
	findBySlug  func(string) (*ChannelDetails, error)
	toggleSubFn func(slug, userId string) (bool, error)
}

func (m *mockChannelRepo) FindAll() *[]Channel { return m.all }
func (m *mockChannelRepo) FindBySlug(slug string) (*ChannelDetails, error) {
	if m.findBySlug != nil {
		return m.findBySlug(slug)
	}
	return m.details, nil
}
func (m *mockChannelRepo) ToggleSubscribe(slug, userId string) (bool, error) {
	return m.toggleSubFn(slug, userId)
}

func newChannelService(repo ChannelRepositoryPort) *ChannelService {
	return NewChannelService(&ChannelServiceDeps{ChannelRepository: repo})
}

func TestGetBySlug(t *testing.T) {
	details := &ChannelDetails{}
	details.Slug = "redgroup"
	svc := newChannelService(&mockChannelRepo{details: details})

	got, err := svc.GetBySlug("redgroup")
	if err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}
	if got.Slug != "redgroup" {
		t.Fatalf("slug = %q", got.Slug)
	}
}

func TestToggleSubscribeMessages(t *testing.T) {
	subSvc := newChannelService(&mockChannelRepo{
		toggleSubFn: func(string, string) (bool, error) { return true, nil },
	})
	res, err := subSvc.ToggleSubscribe("redgroup", "u1")
	if err != nil {
		t.Fatalf("ToggleSubscribe: %v", err)
	}
	if !res.IsSubscribed || res.Message != "Subscribed successfully" {
		t.Fatalf("subscribe result = %+v", res)
	}

	unsubSvc := newChannelService(&mockChannelRepo{
		toggleSubFn: func(string, string) (bool, error) { return false, nil },
	})
	res, err = unsubSvc.ToggleSubscribe("redgroup", "u1")
	if err != nil {
		t.Fatalf("ToggleSubscribe: %v", err)
	}
	if res.IsSubscribed || res.Message != "Unsubscribed successfully" {
		t.Fatalf("unsubscribe result = %+v", res)
	}
}

func TestToggleSubscribeNotFoundMapping(t *testing.T) {
	svc := newChannelService(&mockChannelRepo{
		toggleSubFn: func(string, string) (bool, error) { return false, gorm.ErrRecordNotFound },
	})
	_, err := svc.ToggleSubscribe("missing", "u1")
	if err == nil || err.Error() != ErrChannelNotExist {
		t.Fatalf("err = %v, want %q", err, ErrChannelNotExist)
	}
}

func TestToggleSubscribePropagatesOtherErrors(t *testing.T) {
	boom := errors.New("db down")
	svc := newChannelService(&mockChannelRepo{
		toggleSubFn: func(string, string) (bool, error) { return false, boom },
	})
	if _, err := svc.ToggleSubscribe("x", "u1"); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want the underlying error", err)
	}
}
