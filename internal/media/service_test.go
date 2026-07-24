package media

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"testing"
)

// memStorage is a minimal in-memory ObjectStorage for the media unit tests. It
// is defined locally (not in pkg/testutil) because an in-package media test may
// not import a package that itself imports media.
type memStorage struct {
	mu      sync.Mutex
	objects map[string][]byte
}

func newMemStorage() *memStorage { return &memStorage{objects: map[string][]byte{}} }

func (m *memStorage) Put(_ context.Context, key string, r io.Reader, _ int64, _ string) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.objects[key] = data
	m.mu.Unlock()
	return nil
}

func (m *memStorage) Get(_ context.Context, key string) (io.ReadCloser, ObjectInfo, error) {
	m.mu.Lock()
	data, ok := m.objects[key]
	m.mu.Unlock()
	if !ok {
		return nil, ObjectInfo{}, ErrObjectNotFound
	}
	return io.NopCloser(bytes.NewReader(data)), ObjectInfo{Size: int64(len(data))}, nil
}

func TestGenerateFilenameKeepsExtension(t *testing.T) {
	name := GenerateFilename("clip.MP4")
	if !strings.HasSuffix(name, ".MP4") {
		t.Fatalf("name %q lost its extension", name)
	}
	// 12 hex chars + ".MP4"
	if len(name) != 12+len(".MP4") {
		t.Fatalf("unexpected length for %q", name)
	}
	first := GenerateFilename("a.jpg")
	second := GenerateFilename("a.jpg")
	if first == second {
		t.Fatal("GenerateFilename is not random")
	}
}

func TestValidateFolder(t *testing.T) {
	for _, ok := range []string{"", "avatars", "BANNERS", "thumbnails", "videos"} {
		if err := ValidateFolder(ok); err != nil {
			t.Errorf("ValidateFolder(%q) = %v, want nil", ok, err)
		}
	}
	if err := ValidateFolder("hackers"); err == nil {
		t.Fatal("ValidateFolder accepted an unknown folder")
	}
}

func TestValidateFiles(t *testing.T) {
	if err := ValidateFiles(nil); err == nil {
		t.Fatal("ValidateFiles accepted zero files")
	}
	if err := ValidateFiles([]UploadFile{{MimeType: "image/jpeg", Size: 10}}); err != nil {
		t.Fatalf("ValidateFiles rejected a valid file: %v", err)
	}
	if err := ValidateFiles([]UploadFile{{MimeType: "application/zip", Size: 10}}); err == nil {
		t.Fatal("ValidateFiles accepted an unsupported type")
	}
	if err := ValidateFiles([]UploadFile{{MimeType: "image/png", Size: maxFileSize + 1}}); err == nil {
		t.Fatal("ValidateFiles accepted an oversize file")
	}
}

func TestMapResolution(t *testing.T) {
	cases := []struct {
		w, h int
		want string
	}{
		{3840, 2160, Quality4K},
		{2560, 1440, Quality2K},
		{1920, 1080, Quality1080},
		{1280, 720, Quality720},
		{854, 480, Quality480},
		{640, 360, Quality360},
		{100, 100, Quality720}, // below all thresholds → default
	}
	for _, c := range cases {
		if got := mapResolution(c.w, c.h); got != c.want {
			t.Errorf("mapResolution(%d,%d) = %q, want %q", c.w, c.h, got, c.want)
		}
	}
}

func TestIsVideoAndOriginalName(t *testing.T) {
	if !isVideo(UploadFile{MimeType: "video/mp4"}) {
		t.Error("isVideo said mp4 is not a video")
	}
	if isVideo(UploadFile{MimeType: "image/png"}) {
		t.Error("isVideo said png is a video")
	}
	if originalName(UploadFile{OriginalName: "orig.jpg", Name: "n"}) != "orig.jpg" {
		t.Error("originalName should prefer OriginalName")
	}
	if originalName(UploadFile{Name: "fallback.jpg"}) != "fallback.jpg" {
		t.Error("originalName should fall back to Name")
	}
}

func TestSaveMediaImageUploadsAndReturnsURL(t *testing.T) {
	storage := newMemStorage()
	svc := NewMediaService(&MediaServiceDeps{Storage: storage, PublicBaseURL: "/uploads"})

	files := []UploadFile{{
		OriginalName: "avatar.png",
		MimeType:     "image/png",
		Size:         4,
		Buffer:       []byte("data"),
	}}

	resp, err := svc.SaveMedia(files, "avatars")
	if err != nil {
		t.Fatalf("SaveMedia: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("got %d responses, want 1", len(resp))
	}
	got := resp[0]
	if got.MaxResolution != "" {
		t.Errorf("image response should not carry MaxResolution, got %q", got.MaxResolution)
	}
	wantPrefix := "/uploads/avatars/"
	if !strings.HasPrefix(got.Url, wantPrefix) || !strings.HasSuffix(got.Url, got.Name) {
		t.Errorf("Url = %q, want %s<name>", got.Url, wantPrefix)
	}
	key := "avatars/" + got.Name
	if _, ok := storage.objects[key]; !ok {
		t.Errorf("object %q was not stored; keys=%v", key, storage.objects)
	}
}

func TestSaveMediaDefaultFolder(t *testing.T) {
	storage := newMemStorage()
	svc := NewMediaService(&MediaServiceDeps{Storage: storage, PublicBaseURL: "/uploads"})

	resp, err := svc.SaveMedia([]UploadFile{{
		OriginalName: "x.jpg", MimeType: "image/jpeg", Size: 1, Buffer: []byte("x"),
	}}, "")
	if err != nil {
		t.Fatalf("SaveMedia: %v", err)
	}
	if !strings.HasPrefix(resp[0].Url, "/uploads/default/") {
		t.Fatalf("empty folder should map to default, got %q", resp[0].Url)
	}
}

func TestGetProcessingStatusUnknown(t *testing.T) {
	svc := NewMediaService(&MediaServiceDeps{Storage: newMemStorage(), PublicBaseURL: "/uploads"})
	if s := svc.GetProcessingStatus("nope.mp4"); s != 0 {
		t.Fatalf("unknown file status = %v, want 0", s)
	}
}
