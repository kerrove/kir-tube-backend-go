package media

import (
	"go/kir-tube/configs"
	"go/kir-tube/pkg/di"
	"go/kir-tube/pkg/logs"
	"go/kir-tube/pkg/middleware"
	"net/http"
)

type MediaHandlerDeps struct {
	MediaService *MediaService
	Config       *configs.Config
	UserProvider di.IUserProvider
}
type MediaHandler struct {
	MediaService *MediaService
	Config       *configs.Config
}

func NewMediaHandler(router *http.ServeMux, deps MediaHandlerDeps) {
	handler := &MediaHandler{
		MediaService: deps.MediaService,
		Config:       deps.Config,
	}

	logs.RouteLog(router, "POST /upload-file", middleware.IsAuthed(handler.UploadMediaFile(), deps.Config, deps.UserProvider))
	logs.RouteLog(router, "GET /upload-file/status/{fileName}", middleware.IsAuthed(handler.GetProcessingStatus(), deps.Config, deps.UserProvider))
	logs.RouteLog(router, "GET /uploads", handler.GetFile())
	logs.RouteLog(router, "GET /uploads/index.html", handler.GetFile())
}

func (h *MediaHandler) UploadMediaFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
func (h *MediaHandler) GetFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
func (h *MediaHandler) GetProcessingStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
