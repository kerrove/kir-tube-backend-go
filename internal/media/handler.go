package media

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go/kir-tube/configs"
	"go/kir-tube/pkg/di"
	"go/kir-tube/pkg/logs"
	"go/kir-tube/pkg/middleware"
	"go/kir-tube/pkg/res"
)

const uploadFormField = "file"
const maxUploadMemory = 64 << 20 // 64 MB

type IMediaService interface {
	SaveMedia(files []UploadFile, folder string) ([]MediaResponse, error)
	GetProcessingStatus(fileName string) float64
}

type MediaHandlerDeps struct {
	MediaService IMediaService
	Config       *configs.Config
	UserProvider di.IUserProvider
	OutputDir    string
}
type MediaHandler struct {
	MediaService IMediaService
	Config       *configs.Config
	outputDir    string
}

func NewMediaHandler(router *http.ServeMux, deps MediaHandlerDeps) {
	handler := &MediaHandler{
		MediaService: deps.MediaService,
		Config:       deps.Config,
		outputDir:    deps.OutputDir,
	}

	logs.RouteLog(router, "POST /upload-file", middleware.IsAuthed(handler.UploadMediaFile(), deps.Config, deps.UserProvider))
	logs.RouteLog(router, "GET /upload-file/status/{fileName}", middleware.IsAuthed(handler.GetProcessingStatus(), deps.Config, deps.UserProvider))

	logs.RouteLog(router, "GET /uploads", handler.Forbidden())
	logs.RouteLog(router, "GET /uploads/index.html", handler.Forbidden())
	logs.RouteLog(router, "GET /uploads/{path...}", handler.ServeUploads())
}

func (h *MediaHandler) UploadMediaFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		folder := r.URL.Query().Get("folder")
		if err := ValidateFolder(folder); err != nil {
			writeMediaError(w, err)
			return
		}

		if err := r.ParseMultipartForm(maxUploadMemory); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		files, err := readUploadFiles(r)
		if err != nil {
			writeMediaError(w, err)
			return
		}
		if err := ValidateFiles(files); err != nil {
			writeMediaError(w, err)
			return
		}

		responses, err := h.MediaService.SaveMedia(files, folder)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		res.Json(w, responses, http.StatusOK)
	}
}

func (h *MediaHandler) GetProcessingStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileName := r.PathValue("fileName")
		status := h.MediaService.GetProcessingStatus(fileName)

		res.Json(w, map[string]any{
			"fileName": fileName,
			"status":   status,
		}, http.StatusOK)
	}
}

func (h *MediaHandler) Forbidden() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Access to this directory is forbidden", http.StatusForbidden)
	}
}

func (h *MediaHandler) ServeUploads() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rel := strings.TrimPrefix(r.URL.Path, "/uploads/")

		root, err := filepath.Abs(h.outputDir)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		full := filepath.Join(root, filepath.Clean("/"+rel))

		if full != root && !strings.HasPrefix(full, root+string(os.PathSeparator)) {
			http.Error(w, "Access to this directory is forbidden", http.StatusForbidden)
			return
		}

		info, err := os.Stat(full)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if info.IsDir() {
			http.Error(w, "Access to this directory is forbidden", http.StatusForbidden)
			return
		}

		http.ServeFile(w, r, full)
	}
}

func readUploadFiles(r *http.Request) ([]UploadFile, error) {
	if r.MultipartForm == nil {
		return nil, ErrNoFile
	}

	headers := r.MultipartForm.File[uploadFormField]
	files := make([]UploadFile, 0, len(headers))

	for _, header := range headers {
		opened, err := header.Open()
		if err != nil {
			return nil, err
		}

		buffer, err := io.ReadAll(opened)
		opened.Close()
		if err != nil {
			return nil, err
		}

		files = append(files, UploadFile{
			Name:         header.Filename,
			OriginalName: header.Filename,
			MimeType:     header.Header.Get("Content-Type"),
			Size:         header.Size,
			Buffer:       buffer,
		})
	}

	return files, nil
}

func writeMediaError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNoFile),
		errors.Is(err, ErrUnsupportedType),
		errors.Is(err, ErrFileTooBig),
		errors.Is(err, ErrInvalidFolder):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
