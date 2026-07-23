package studio

import (
	"errors"
	"go/kir-tube/configs"
	"go/kir-tube/internal/video"
	"go/kir-tube/pkg/di"
	"go/kir-tube/pkg/logs"
	"go/kir-tube/pkg/middleware"
	request "go/kir-tube/pkg/req"
	"go/kir-tube/pkg/res"
	"net/http"

	"gorm.io/gorm"
)

type StudioHandlerDeps struct {
	StudioService *StudioService
	Config        *configs.Config
	UserProvider  di.IUserProvider
}
type StudioHandler struct {
	StudioService *StudioService
	Config        *configs.Config
}

func NewStudioHandler(router *http.ServeMux, deps StudioHandlerDeps) {
	handler := &StudioHandler{
		StudioService: deps.StudioService,
		Config:        deps.Config,
	}

	logs.RouteLog(router, "GET /studio/videos", middleware.IsAuthed(handler.GetAll(), deps.Config, deps.UserProvider))
	logs.RouteLog(router, "GET /studio/videos/{id}", middleware.IsAuthed(handler.GetById(), deps.Config, deps.UserProvider))
	logs.RouteLog(router, "POST /studio/videos", middleware.IsAuthed(handler.Create(), deps.Config, deps.UserProvider))
	logs.RouteLog(router, "PUT /studio/videos/{id}", middleware.IsAuthed(handler.Update(), deps.Config, deps.UserProvider))
	logs.RouteLog(router, "DELETE /studio/videos/{id}", middleware.IsAuthed(handler.Delete(), deps.Config, deps.UserProvider))
}

func (h *StudioHandler) GetAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(middleware.ContextUserKey).(*di.ContextUser)
		searchTerm := r.URL.Query().Get("searchTerm")

		page, limit := video.PaginationParams(r)

		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		if user.Channel == nil {
			http.Error(w, "user has no channel", http.StatusForbidden)
			return
		}

		videos, err := h.StudioService.GetAll(user.Channel.ID, searchTerm, page, limit)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		res.Json(w, videos, http.StatusOK)
	}
}
func (h *StudioHandler) GetById() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		video, err := h.StudioService.ById(id)
		if err != nil {
			writeServiceError(w, err)
			return
		}

		res.Json(w, video, http.StatusOK)
	}
}
func (h *StudioHandler) Create() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := request.HandleBody[video.CreateVideoInput](&w, r)

		if err != nil {
			return
		}
		user, ok := r.Context().Value(middleware.ContextUserKey).(*di.ContextUser)
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		if user.Channel == nil {
			http.Error(w, "user has no channel", http.StatusForbidden)
			return
		}

		video, err := h.StudioService.Create(user.Channel.ID, *body)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		res.Json(w, video, http.StatusOK)
	}
}
func (h *StudioHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := request.HandleBody[video.UpdateVideoInput](&w, r)

		if err != nil {
			return
		}
		id := r.PathValue("id")

		video, err := h.StudioService.Update(id, *body)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		res.Json(w, video, http.StatusOK)
	}
}
func (h *StudioHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		video, err := h.StudioService.Delete(id)
		if err != nil {
			writeServiceError(w, err)
			return
		}
		res.Json(w, video, http.StatusOK)
	}
}

func writeServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, video.ErrVideoNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
