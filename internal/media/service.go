package media

import (
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
	OutputDir     string
	PublicBaseURL string
}

type MediaService struct {
	outputDir        string
	publicBaseURL    string
	mu               sync.RWMutex
	processingStatus map[string]float64
}

func NewMediaService(deps *MediaServiceDeps) *MediaService {
	return &MediaService{
		outputDir:        deps.OutputDir,
		publicBaseURL:    strings.TrimRight(deps.PublicBaseURL, "/"),
		processingStatus: make(map[string]float64),
	}
}

func (s *MediaService) SaveMedia(files []UploadFile, folder string) ([]MediaResponse, error) {
	if folder == "" {
		folder = "default"
	}
	folder = strings.ToLower(folder)

	uploadFolder := filepath.Join(s.outputDir, folder)
	if err := os.MkdirAll(uploadFolder, 0o755); err != nil {
		return nil, err
	}

	file := files[0]

	uniqueName := GenerateFilename(originalName(file))
	filePath := filepath.Join(uploadFolder, uniqueName)

	if err := os.WriteFile(filePath, file.Buffer, 0o644); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s/%s", s.publicBaseURL, folder, uniqueName)

	if !isVideo(file) {
		return []MediaResponse{{Url: url, Name: uniqueName}}, nil
	}

	width, height, err := s.getVideoResolution(filePath)
	if err != nil {
		return nil, err
	}
	maxResolution := mapResolution(width, height)

	s.setStatus(uniqueName, statusQueued)

	go func() {
		if err := s.processVideo(filePath, uniqueName, folder); err != nil {
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
// it, updating the status after each one completes.
func (s *MediaService) processVideo(inputPath, fileName, folder string) error {
	width, height, err := s.getVideoResolution(inputPath)
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
		if err := s.convertVideo(inputPath, resolution, fileName, folder); err != nil {
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

// convertVideo runs ffmpeg to resize the source into OutputDir/<folder>/<name>/<file>.
func (s *MediaService) convertVideo(inputPath string, resolution Resolution, fileName, folder string) error {
	outputDir := filepath.Join(s.outputDir, folder, resolution.Name)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	outputPath := filepath.Join(outputDir, fileName)

	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-i", inputPath,
		"-s", fmt.Sprintf("%dx%d", resolution.Width, resolution.Height),
		outputPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg %s: %w: %s", resolution.Name, err, out)
	}
	return nil
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
