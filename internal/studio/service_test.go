package studio

import (
	"errors"
	"testing"

	"gorm.io/gorm"

	"go/kir-tube/internal/video"
)

type mockVideoRepo struct {
	findChannel  func(channelID, term string, skip, limit int) ([]video.Video, error)
	countChannel func(channelID, term string) (int64, error)
	findById     func(string) (*video.Video, error)
	createFn     func(channelID string, input video.CreateVideoInput) (*video.Video, error)
	updateFn     func(id string, input video.UpdateVideoInput) (*video.Video, error)
	deleteFn     func(string) (*video.Video, error)
	lastSkip     int
	lastLimit    int
}

func (m *mockVideoRepo) FindChannelVideos(c, term string, skip, limit int) ([]video.Video, error) {
	m.lastSkip, m.lastLimit = skip, limit
	if m.findChannel != nil {
		return m.findChannel(c, term, skip, limit)
	}
	return nil, nil
}
func (m *mockVideoRepo) CountChannelVideos(c, term string) (int64, error) {
	if m.countChannel != nil {
		return m.countChannel(c, term)
	}
	return 0, nil
}
func (m *mockVideoRepo) FindById(id string) (*video.Video, error) { return m.findById(id) }
func (m *mockVideoRepo) Create(c string, in video.CreateVideoInput) (*video.Video, error) {
	return m.createFn(c, in)
}
func (m *mockVideoRepo) Update(id string, in video.UpdateVideoInput) (*video.Video, error) {
	return m.updateFn(id, in)
}
func (m *mockVideoRepo) Delete(id string) (*video.Video, error) { return m.deleteFn(id) }

func newStudioService(repo VideoRepository) *StudioService {
	return NewStudioService(&StudioServiceDeps{VideoRepository: repo})
}

func TestGetAllPaginationDefaults(t *testing.T) {
	repo := &mockVideoRepo{
		countChannel: func(string, string) (int64, error) { return 13, nil },
	}
	svc := newStudioService(repo)

	// page/limit below 1 fall back to page 1 and defaultStudioLimit (6).
	res, err := svc.GetAll("c1", "", 0, 0)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if res.Page != 1 || res.Limit != defaultStudioLimit {
		t.Fatalf("page/limit = %d/%d", res.Page, res.Limit)
	}
	if repo.lastSkip != 0 {
		t.Fatalf("skip = %d, want 0", repo.lastSkip)
	}
	// ceil(13/6) = 3
	if res.TotalPages != 3 {
		t.Fatalf("totalPages = %d, want 3", res.TotalPages)
	}
}

func TestGetAllSkipCalculation(t *testing.T) {
	repo := &mockVideoRepo{}
	svc := newStudioService(repo)
	if _, err := svc.GetAll("c1", "", 3, 10); err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if repo.lastSkip != 20 || repo.lastLimit != 10 {
		t.Fatalf("skip/limit = %d/%d, want 20/10", repo.lastSkip, repo.lastLimit)
	}
}

func TestCreateReturnsID(t *testing.T) {
	created := &video.Video{}
	created.ID = "vid-123"
	svc := newStudioService(&mockVideoRepo{
		createFn: func(string, video.CreateVideoInput) (*video.Video, error) { return created, nil },
	})
	id, err := svc.Create("c1", video.CreateVideoInput{Title: "T"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != "vid-123" {
		t.Fatalf("id = %q", id)
	}
}

func TestByIdNotFoundMapping(t *testing.T) {
	svc := newStudioService(&mockVideoRepo{
		findById: func(string) (*video.Video, error) { return nil, gorm.ErrRecordNotFound },
	})
	_, err := svc.ById("missing")
	if !errors.Is(err, video.ErrVideoNotFound) {
		t.Fatalf("err = %v, want video.ErrVideoNotFound", err)
	}
}

func TestDeleteNotFoundMapping(t *testing.T) {
	svc := newStudioService(&mockVideoRepo{
		deleteFn: func(string) (*video.Video, error) { return nil, gorm.ErrRecordNotFound },
	})
	if _, err := svc.Delete("missing"); !errors.Is(err, video.ErrVideoNotFound) {
		t.Fatalf("err = %v, want video.ErrVideoNotFound", err)
	}
}

func TestUpdatePropagatesOtherErrors(t *testing.T) {
	boom := errors.New("db down")
	svc := newStudioService(&mockVideoRepo{
		updateFn: func(string, video.UpdateVideoInput) (*video.Video, error) { return nil, boom },
	})
	if _, err := svc.Update("v1", video.UpdateVideoInput{}); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}
