package video

import (
	"errors"
	"testing"
	"time"
)

// stubRepo embeds IVideoRepository so it satisfies the full contract; only the
// methods a given test needs are overridden. Any unstubbed call panics (nil
// interface), which surfaces accidental dependencies.
type stubRepo struct {
	IVideoRepository

	findPublic     func(term string, skip, limit int) ([]Video, error)
	countPublic    func(term string) (int64, error)
	findByChannel  func(channelID string, skip, limit int) ([]Video, error)
	countByChannel func(channelID string) (int64, error)
	incrementViews func(publicId string) (*Video, error)
	toggleLike     func(userId, videoId string) (bool, error)
	findRecent     func(since time.Time) ([]Video, error)
	findByViews    func(exclude []string, limit int) ([]Video, error)
	lastSkip       int
	lastLimit      int
}

func (s *stubRepo) FindPublicVideos(term string, skip, limit int) ([]Video, error) {
	s.lastSkip, s.lastLimit = skip, limit
	return s.findPublic(term, skip, limit)
}
func (s *stubRepo) CountPublicVideos(term string) (int64, error) { return s.countPublic(term) }
func (s *stubRepo) FindByChannel(c string, skip, limit int) ([]Video, error) {
	s.lastSkip, s.lastLimit = skip, limit
	return s.findByChannel(c, skip, limit)
}
func (s *stubRepo) CountByChannel(c string) (int64, error)              { return s.countByChannel(c) }
func (s *stubRepo) IncrementViewsCount(p string) (*Video, error)        { return s.incrementViews(p) }
func (s *stubRepo) ToggleLike(u, v string) (bool, error)                { return s.toggleLike(u, v) }
func (s *stubRepo) FindRecentPublicVideos(t time.Time) ([]Video, error) { return s.findRecent(t) }
func (s *stubRepo) FindPublicVideosByViews(e []string, l int) ([]Video, error) {
	return s.findByViews(e, l)
}

func newVideoService(r IVideoRepository) *VideoService {
	return NewVideoService(&VideoServiceDeps{VideoRepository: r})
}

func TestGetAllPaginationAndTotals(t *testing.T) {
	repo := &stubRepo{
		findPublic:  func(string, int, int) ([]Video, error) { return []Video{{}, {}}, nil },
		countPublic: func(string) (int64, error) { return 25, nil },
	}
	svc := newVideoService(repo)

	// page/limit 0 → defaults page 1, limit 10.
	res, err := svc.GetAll("", 0, 0)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if res.Page != 1 || res.Limit != 10 || repo.lastSkip != 0 {
		t.Fatalf("defaults wrong: page=%d limit=%d skip=%d", res.Page, res.Limit, repo.lastSkip)
	}
	// ceil(25/10) = 3
	if res.TotalPages != 3 || res.TotalCount != 25 {
		t.Fatalf("totals wrong: pages=%d count=%d", res.TotalPages, res.TotalCount)
	}
}

func TestGetAllSkip(t *testing.T) {
	repo := &stubRepo{
		findPublic:  func(string, int, int) ([]Video, error) { return nil, nil },
		countPublic: func(string) (int64, error) { return 0, nil },
	}
	svc := newVideoService(repo)
	if _, err := svc.GetAll("q", 3, 5); err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if repo.lastSkip != 10 || repo.lastLimit != 5 {
		t.Fatalf("skip/limit = %d/%d, want 10/5", repo.lastSkip, repo.lastLimit)
	}
}

func TestGetAllPropagatesError(t *testing.T) {
	boom := errors.New("db down")
	repo := &stubRepo{
		findPublic: func(string, int, int) ([]Video, error) { return nil, boom },
	}
	if _, err := newVideoService(repo).GetAll("", 1, 10); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}

func TestByChannel(t *testing.T) {
	repo := &stubRepo{
		findByChannel:  func(string, int, int) ([]Video, error) { return []Video{{}}, nil },
		countByChannel: func(string) (int64, error) { return 1, nil },
	}
	res, err := newVideoService(repo).ByChannel("c1", 2, 10)
	if err != nil {
		t.Fatalf("ByChannel: %v", err)
	}
	if repo.lastSkip != 10 || res.TotalCount != 1 {
		t.Fatalf("unexpected: skip=%d count=%d", repo.lastSkip, res.TotalCount)
	}
}

func TestUpdateViewsCountDelegates(t *testing.T) {
	want := &Video{}
	want.ID = "v1"
	repo := &stubRepo{incrementViews: func(string) (*Video, error) { return want, nil }}
	got, err := newVideoService(repo).UpdateViewsCount("public")
	if err != nil {
		t.Fatalf("UpdateViewsCount: %v", err)
	}
	if got.ID != "v1" {
		t.Fatalf("id = %q", got.ID)
	}
}

func TestToggleLikeDelegates(t *testing.T) {
	repo := &stubRepo{toggleLike: func(string, string) (bool, error) { return true, nil }}
	liked, err := newVideoService(repo).ToggleLike("u1", "v1")
	if err != nil {
		t.Fatalf("ToggleLike: %v", err)
	}
	if !liked {
		t.Fatal("expected liked=true")
	}
}

func TestGetTrendingVideosSortsByEngagement(t *testing.T) {
	now := time.Now()
	low := Video{ViewsCount: 1}
	low.ID, low.CreatedAt = "low", now
	high := Video{ViewsCount: 1000}
	high.ID, high.CreatedAt = "high", now

	repo := &stubRepo{
		// return low first; the service must reorder by engagement (views).
		findRecent:  func(time.Time) ([]Video, error) { return []Video{low, high}, nil },
		findByViews: func([]string, int) ([]Video, error) { return nil, nil },
	}
	trending, err := newVideoService(repo).GetTrendingVideos()
	if err != nil {
		t.Fatalf("GetTrendingVideos: %v", err)
	}
	if len(trending) < 2 || trending[0].ID != "high" {
		t.Fatalf("expected 'high' first, got %+v", trending)
	}
}

func TestGetTrendingVideosBackfills(t *testing.T) {
	now := time.Now()
	only := Video{ViewsCount: 5}
	only.ID, only.CreatedAt = "only", now

	backfill := make([]Video, 5)
	for i := range backfill {
		backfill[i].ID = "bf"
	}

	repo := &stubRepo{
		findRecent:  func(time.Time) ([]Video, error) { return []Video{only}, nil },
		findByViews: func(_ []string, limit int) ([]Video, error) { return backfill[:limit], nil },
	}
	trending, err := newVideoService(repo).GetTrendingVideos()
	if err != nil {
		t.Fatalf("GetTrendingVideos: %v", err)
	}
	if len(trending) != 6 {
		t.Fatalf("expected 6 after backfill, got %d", len(trending))
	}
}
