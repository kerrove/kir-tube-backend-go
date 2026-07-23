package video

import (
	"net/http"
	"strings"

	"go/kir-tube/configs"
	"go/kir-tube/pkg/di"
	"go/kir-tube/pkg/logs"
	"go/kir-tube/pkg/middleware"
	request "go/kir-tube/pkg/req"
	"go/kir-tube/pkg/res"
)

type VideoHandlerDeps struct {
	VideoService *VideoService
	Config       *configs.Config
	UserProvider di.IUserProvider
}
type VideoHandler struct {
	VideoService *VideoService
	Config       *configs.Config
}

func NewVideoHandler(router *http.ServeMux, deps VideoHandlerDeps) {
	handler := &VideoHandler{
		VideoService: deps.VideoService,
		Config:       deps.Config,
	}

	logs.RouteLog(router, "POST /users/profile/likes", middleware.IsAuthed(handler.ToggleLike(), deps.Config, deps.UserProvider))

	logs.RouteLog(router, "GET /videos/by-publicId/{publicId}", handler.GetByPublicId())
	logs.RouteLog(router, "GET /videos/by-channel/{channelId}", handler.GetByChannel())
	logs.RouteLog(router, "GET /videos", handler.GetAll())
	logs.RouteLog(router, "GET /videos/games", handler.GetGames())
	logs.RouteLog(router, "GET /videos/trending", handler.GetTrending())
	logs.RouteLog(router, "GET /videos/explore", middleware.MaybeAuthed(handler.GetExplore(), deps.Config, deps.UserProvider))
	logs.RouteLog(router, "PUT /videos/update-views-count/{publicId}", handler.UpdateViewsCount())
}

func (h *VideoHandler) GetByPublicId() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("publicId")
		video, err := h.VideoService.GetVideoByPublicId(id)

		if err != nil {
			res.WriteServiceError(w, err, ErrVideoNotFound)
			return
		}

		res.Json(w, video, http.StatusOK)
	}
}
func (h *VideoHandler) GetTrending() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		videos, err := h.VideoService.GetTrendingVideos()

		if err != nil {
			res.WriteServiceError(w, err, ErrVideoNotFound)
			return
		}

		res.Json(w, videos, http.StatusOK)
	}
}
func (h *VideoHandler) GetAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		searchTerm := r.URL.Query().Get("searchTerm")

		page, limit := PaginationParams(r)

		videos, err := h.VideoService.GetAll(searchTerm, page, limit)

		if err != nil {
			res.WriteServiceError(w, err, ErrVideoNotFound)
			return
		}

		res.Json(w, videos, http.StatusOK)
	}
}
func (h *VideoHandler) UpdateViewsCount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		publicId := r.PathValue("publicId")

		video, err := h.VideoService.UpdateViewsCount(publicId)

		if err != nil {
			res.WriteServiceError(w, err, ErrVideoNotFound)
			return
		}

		res.Json(w, video, http.StatusOK)
	}
}
func (h *VideoHandler) GetByChannel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channelId := r.PathValue("channelId")
		page, limit := PaginationParams(r)

		videos, err := h.VideoService.ByChannel(channelId, page, limit)

		if err != nil {
			res.WriteServiceError(w, err, ErrVideoNotFound)
			return
		}

		res.Json(w, videos, http.StatusOK)
	}
}
func (h *VideoHandler) GetExplore() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := request.GetProfileIdOptional(r)

		var excludeIds []string
		if raw := r.URL.Query().Get("excludeIds"); raw != "" {
			excludeIds = strings.Split(raw, ",")
		}

		page, limit := PaginationParams(r)

		videos, err := h.VideoService.GetRecommendations(id, page, limit, excludeIds)

		if err != nil {
			res.WriteServiceError(w, err, ErrVideoNotFound)
			return
		}

		res.Json(w, videos, http.StatusOK)
	}
}

func (h *VideoHandler) GetGames() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		videos, err := h.VideoService.GetTrendingVideos()

		if err != nil {
			res.WriteServiceError(w, err, ErrVideoNotFound)
			return
		}

		res.Json(w, videos, http.StatusOK)
	}
}
func (h *VideoHandler) ToggleLike() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := request.GetProfileId(w, r)
		body, err := request.HandleBody[ToggleLikeReq](&w, r)

		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		liked, err := h.VideoService.ToggleLike(id, body.VideoId)

		if err != nil {
			res.WriteServiceError(w, err, ErrVideoNotFound)
			return
		}

		res.Json(w, map[string]bool{"liked": liked}, http.StatusOK)
	}
}
