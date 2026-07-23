package playlist

import (
	"go/kir-tube/configs"
	"go/kir-tube/pkg/di"
	"go/kir-tube/pkg/logs"
	"go/kir-tube/pkg/middleware"
	request "go/kir-tube/pkg/req"
	"go/kir-tube/pkg/res"
	"net/http"
)

type PlaylistHandlerDeps struct {
	PlaylistService *PlaylistService
	Config          *configs.Config
	UserProvider    di.IUserProvider
}
type PlaylistHandler struct {
	PlaylistService *PlaylistService
	Config          *configs.Config
}

func NewPlaylistHandler(router *http.ServeMux, deps PlaylistHandlerDeps) {
	handler := &PlaylistHandler{
		PlaylistService: deps.PlaylistService,
		Config:          deps.Config,
	}

	logs.RouteLog(router, "GET /playlists", middleware.IsAuthed(handler.GetUserPlaylist(), deps.Config, deps.UserProvider))
	logs.RouteLog(router, "GET /playlists/{playlistId}", middleware.IsAuthed(handler.GetPlaylistById(), deps.Config, deps.UserProvider))
	logs.RouteLog(router, "POST /playlists/{playlistId}/toggle-video", middleware.IsAuthed(handler.ToggleVideo(), deps.Config, deps.UserProvider))
	logs.RouteLog(router, "POST /playlists", middleware.IsAuthed(handler.Create(), deps.Config, deps.UserProvider))
}

func (h *PlaylistHandler) GetUserPlaylist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := request.GetProfileId(w, r)

		playlist, err := h.PlaylistService.GetUserPlaylist(userId)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		res.Json(w, playlist, http.StatusOK)
	}
}
func (h *PlaylistHandler) GetPlaylistById() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playlistId := r.PathValue("playlistId")

		playlist, err := h.PlaylistService.GetPlaylistById(playlistId)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		res.Json(w, playlist, http.StatusOK)
	}
}
func (h *PlaylistHandler) ToggleVideo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playlistId := r.PathValue("playlistId")
		userId := request.GetProfileId(w, r)

		body, err := request.HandleBody[ToggleVideoRequest](&w, r)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		result, err := h.PlaylistService.ToggleVideo(userId, playlistId, body.VideoId)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		res.Json(w, result, http.StatusOK)
	}
}
func (h *PlaylistHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := request.GetProfileId(w, r)

		body, err := request.HandleBody[PlaylistRequest](&w, r)
		if err != nil {
			return
		}

		playlist, err := h.PlaylistService.Create(userId, body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		res.Json(w, playlist, http.StatusOK)

	}
}
