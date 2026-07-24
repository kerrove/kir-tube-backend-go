package comment

import (
	"errors"
	"testing"

	"go/kir-tube/internal/video"
	"go/kir-tube/pkg/di"
)

type mockCommentRepo struct {
	createFn func(userId string, body *CreateCommentReq) (*video.VideoComment, error)
	updateFn func(commentId, userId string, body *UpdateCommentReq) (*video.VideoComment, error)
	deleteFn func(commentId, userId string) (bool, error)
	findFn   func(publicId string) ([]CommentWithUser, error)
}

func (m *mockCommentRepo) Create(u string, b *CreateCommentReq) (*video.VideoComment, error) {
	return m.createFn(u, b)
}
func (m *mockCommentRepo) Update(c, u string, b *UpdateCommentReq) (*video.VideoComment, error) {
	return m.updateFn(c, u, b)
}
func (m *mockCommentRepo) Delete(c, u string) (bool, error)                { return m.deleteFn(c, u) }
func (m *mockCommentRepo) FindByVideo(p string) ([]CommentWithUser, error) { return m.findFn(p) }

// unusedVideoRepo satisfies di.IVideoRepository; the comment service does not
// call it in these tests.
type unusedVideoRepo struct{}

func (unusedVideoRepo) FindSubscribedVideos(string) ([]di.SubscribedVideo, error) { return nil, nil }
func (unusedVideoRepo) FindLikedVideos(string) ([]di.SubscribedVideo, error)      { return nil, nil }
func (unusedVideoRepo) FindSubscriptions(string) ([]di.SubscriptionChannel, error) {
	return nil, nil
}

func newCommentService(repo CommentRepositoryPort) *CommentService {
	return NewCommentService(&CommentServiceDeps{CommentRepository: repo, VideoRepository: unusedVideoRepo{}})
}

func TestCreateCommentDelegates(t *testing.T) {
	want := &video.VideoComment{Text: "hi"}
	var gotUser string
	svc := newCommentService(&mockCommentRepo{
		createFn: func(u string, _ *CreateCommentReq) (*video.VideoComment, error) {
			gotUser = u
			return want, nil
		},
	})
	txt, vid := "hi", "v1"
	got, err := svc.CreateComment("u1", &CreateCommentReq{Text: &txt, VideoId: &vid})
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}
	if got != want || gotUser != "u1" {
		t.Fatalf("delegate failed: user=%q got=%+v", gotUser, got)
	}
}

func TestDeleteCommentPropagatesError(t *testing.T) {
	boom := errors.New("forbidden")
	svc := newCommentService(&mockCommentRepo{
		deleteFn: func(string, string) (bool, error) { return false, boom },
	})
	if _, err := svc.DeleteComment("c1", "u1"); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}

func TestGetByVideoDelegates(t *testing.T) {
	want := []CommentWithUser{{}}
	svc := newCommentService(&mockCommentRepo{
		findFn: func(string) ([]CommentWithUser, error) { return want, nil },
	})
	got, err := svc.GetByVideo("public-id")
	if err != nil {
		t.Fatalf("GetByVideo: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d comments, want 1", len(got))
	}
}
