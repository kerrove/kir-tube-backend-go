package media

import (
	"errors"
	"io"
	"net/http"
	"path"
	"strconv"
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
	Storage      ObjectStorage
}
type MediaHandler struct {
	MediaService IMediaService
	Config       *configs.Config
	storage      ObjectStorage
}

// NewMediaHandler wires the media routes. The upload/status endpoints are API
// routes registered on apiRouter (served under the global /api prefix), while
// the static file serving is registered on publicRouter (the root mux) so media
// is reachable at /uploads/... rather than /api/uploads/....
func NewMediaHandler(apiRouter, publicRouter *http.ServeMux, deps MediaHandlerDeps) {
	handler := &MediaHandler{
		MediaService: deps.MediaService,
		Config:       deps.Config,
		storage:      deps.Storage,
	}

	logs.RouteLog(apiRouter, "POST /upload-file", middleware.IsAuthed(handler.UploadMediaFile(), deps.Config, deps.UserProvider))
	logs.RouteLog(apiRouter, "GET /upload-file/status/{fileName}", middleware.IsAuthed(handler.GetProcessingStatus(), deps.Config, deps.UserProvider))

	// Public file serving at the root: /uploads/...
	handler.registerServeRoutes(publicRouter)
	// Backward-compatible alias for rows persisted before the split, whose URLs
	// still point at /api/uploads/... (apiRouter is mounted under /api).
	handler.registerServeRoutes(apiRouter)
}

func (h *MediaHandler) registerServeRoutes(router *http.ServeMux) {
	logs.RouteLog(router, "GET /uploads", h.Forbidden())
	logs.RouteLog(router, "GET /uploads/index.html", h.Forbidden())
	logs.RouteLog(router, "GET /uploads/{path...}", h.ServeUploads())
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

		// Normalize into a clean, bucket-relative key and reject traversal.
		key := strings.TrimPrefix(path.Clean("/"+rel), "/")
		if key == "" || strings.HasPrefix(key, "../") || strings.Contains(key, "/../") {
			http.Error(w, "Access to this directory is forbidden", http.StatusForbidden)
			return
		}

		obj, info, err := h.storage.Get(r.Context(), key)
		if err != nil {
			if errors.Is(err, ErrObjectNotFound) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer obj.Close()

		if info.ContentType != "" {
			w.Header().Set("Content-Type", info.ContentType)
		}
		if info.Size > 0 {
			w.Header().Set("Content-Length", strconv.FormatInt(info.Size, 10))
		}

		if _, err := io.Copy(w, obj); err != nil {
			// Response is likely already partially written; nothing to recover.
			return
		}
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
