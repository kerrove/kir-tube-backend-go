package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// Processing-status sentinels stored in the status map.
const (
	statusQueued = 0.0   // accepted, not started
	statusDone   = 100.0 // all resolutions transcoded
	statusFailed = -1.0  // transcoding failed
)

type MediaServiceDeps struct {
	Storage       ObjectStorage
	PublicBaseURL string
}

type MediaService struct {
	storage          ObjectStorage
	publicBaseURL    string
	mu               sync.RWMutex
	processingStatus map[string]float64
}

func NewMediaService(deps *MediaServiceDeps) *MediaService {
	return &MediaService{
		storage:          deps.Storage,
		publicBaseURL:    strings.TrimRight(deps.PublicBaseURL, "/"),
		processingStatus: make(map[string]float64),
	}
}

func (s *MediaService) SaveMedia(files []UploadFile, folder string) ([]MediaResponse, error) {
	if folder == "" {
		folder = "default"
	}
	folder = strings.ToLower(folder)

	file := files[0]

	uniqueName := GenerateFilename(originalName(file))
	key := folder + "/" + uniqueName

	if err := s.storage.Put(
		context.Background(),
		key,
		bytes.NewReader(file.Buffer),
		file.Size,
		file.MimeType,
	); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s/%s", s.publicBaseURL, folder, uniqueName)

	if !isVideo(file) {
		return []MediaResponse{{Url: url, Name: uniqueName}}, nil
	}

	// ffmpeg/ffprobe operate on local paths, so stage the source in a temp file.
	// The transcoding goroutine removes it when done.
	srcPath, err := writeTempFile(uniqueName, file.Buffer)
	if err != nil {
		return nil, err
	}

	width, height, err := s.getVideoResolution(srcPath)
	if err != nil {
		os.Remove(srcPath)
		return nil, err
	}
	maxResolution := mapResolution(width, height)

	s.setStatus(uniqueName, statusQueued)

	go func() {
		defer os.Remove(srcPath)
		if err := s.processVideo(srcPath, uniqueName, folder, file.MimeType); err != nil {
			s.setStatus(uniqueName, statusFailed)
			log.Printf("media: video processing failed: %v", err)
			return
		}
		s.setStatus(uniqueName, statusDone)
	}()

	return []MediaResponse{{Url: url, Name: uniqueName, MaxResolution: maxResolution}}, nil
}

// GetProcessingStatus reports transcoding progress for a file: 0 when unknown or
// queued, 100 when done, -1 on failure.
func (s *MediaService) GetProcessingStatus(fileName string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.processingStatus[fileName]
}

func (s *MediaService) setStatus(fileName string, value float64) {
	s.mu.Lock()
	s.processingStatus[fileName] = value
	s.mu.Unlock()
}

// processVideo transcodes the source into every resolution that does not exceed
// it, uploading each variant to MinIO and updating the status after each one.
func (s *MediaService) processVideo(srcPath, fileName, folder, contentType string) error {
	width, height, err := s.getVideoResolution(srcPath)
	if err != nil {
		return err
	}

	var targets []Resolution
	for _, r := range Resolutions {
		if r.Width <= width && r.Height <= height {
			targets = append(targets, r)
		}
	}

	for i, resolution := range targets {
		if err := s.convertVideo(srcPath, resolution, fileName, folder, contentType); err != nil {
			return err
		}
		// Coarser than the original per-frame progress: ffmpeg's stderr would
		// have to be parsed live for that. Status advances once each resolution
		// finishes, which is enough to drive a progress bar.
		s.setStatus(fileName, float64(i+1)/float64(len(targets))*100)
	}

	s.setStatus(fileName, statusDone)
	return nil
}

// convertVideo runs ffmpeg to resize the source into a temp file, then uploads it
// to MinIO under <folder>/<resolution>/<file>.
func (s *MediaService) convertVideo(srcPath string, resolution Resolution, fileName, folder, contentType string) error {
	outPath := srcPath + "." + resolution.Name + filepath.Ext(fileName)
	defer os.Remove(outPath)

	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-i", srcPath,
		"-s", fmt.Sprintf("%dx%d", resolution.Width, resolution.Height),
		outPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg %s: %w: %s", resolution.Name, err, out)
	}

	f, err := os.Open(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	key := folder + "/" + resolution.Name + "/" + fileName
	return s.storage.Put(context.Background(), key, f, info.Size(), contentType)
}

// ffprobeResult is the slice of ffprobe -show_entries JSON we care about.
type ffprobeResult struct {
	Streams []struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"streams"`
}

// getVideoResolution reads the first video stream's dimensions via ffprobe.
func (s *MediaService) getVideoResolution(filePath string) (width, height int, err error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "json",
		filePath,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("ffprobe: %w", err)
	}

	var result ffprobeResult
	if err := json.Unmarshal(out, &result); err != nil {
		return 0, 0, err
	}
	if len(result.Streams) == 0 {
		return 0, 0, fmt.Errorf("media: no video stream found")
	}

	return result.Streams[0].Width, result.Streams[0].Height, nil
}

// writeTempFile stages an upload buffer on local disk (ffmpeg/ffprobe need a
// path). The caller is responsible for removing the returned file.
func writeTempFile(name string, buf []byte) (string, error) {
	f, err := os.CreateTemp("", "kir-tube-*"+filepath.Ext(name))
	if err != nil {
		return "", err
	}
	path := f.Name()

	if _, err := f.Write(buf); err != nil {
		f.Close()
		os.Remove(path)
		return "", err
	}
	if err := f.Close(); err != nil {
		os.Remove(path)
		return "", err
	}
	return path, nil
}

// isVideo reports whether a file is a video by its mime type.
func isVideo(file UploadFile) bool {
	return strings.HasPrefix(file.MimeType, "video/")
}

// originalName returns the best available source name for extension detection.
func originalName(file UploadFile) string {
	if file.OriginalName != "" {
		return file.OriginalName
	}
	return file.Name
}

// mapResolution picks the highest player-quality label the given dimensions
// reach, defaulting to 720p. Mirrors MediaService.mapResolution.
func mapResolution(width, height int) string {
	steps := []Resolution{
		{Name: Quality4K, Width: 3840, Height: 2160},
		{Name: Quality2K, Width: 2560, Height: 1440},
		{Name: Quality1080, Width: 1920, Height: 1080},
		{Name: Quality720, Width: 1280, Height: 720},
		{Name: Quality480, Width: 854, Height: 480},
		{Name: Quality360, Width: 640, Height: 360},
	}
	for _, r := range steps {
		if width >= r.Width && height >= r.Height {
			return r.Name
		}
	}
	return Quality720
}
