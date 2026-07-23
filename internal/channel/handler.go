package channel

import (
	"go/kir-tube/configs"
	"go/kir-tube/pkg/di"
	"go/kir-tube/pkg/logs"
	"go/kir-tube/pkg/middleware"
	request "go/kir-tube/pkg/req"
	"go/kir-tube/pkg/res"
	"net/http"
)

type ChannelHandlerDeps struct {
	*configs.Config
	*ChannelService
	UserProvider di.IUserProvider
}
type ChannelHandler struct {
	*ChannelService
	*configs.Config
}

func NewChannelHandler(router *http.ServeMux, deps ChannelHandlerDeps) {
	handler := &ChannelHandler{
		Config:         deps.Config,
		ChannelService: deps.ChannelService,
	}

	logs.RouteLog(router, "GET /channels", handler.GetAll())
	logs.RouteLog(router, "GET /channels/by-slug/{slug}", handler.GetBySlug())
	logs.RouteLog(router, "PATCH /channels/toggle-subscribe/{slug}", middleware.IsAuthed(handler.ToggleSubscribe(), deps.Config, deps.UserProvider))
}

func (h *ChannelHandler) GetAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channels := h.ChannelService.GetAll()
		res.Json(w, channels, http.StatusOK)
	}
}
func (h *ChannelHandler) GetBySlug() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")

		channel, err := h.ChannelService.GetBySlug(slug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		res.Json(w, channel, http.StatusOK)
	}
}
func (h *ChannelHandler) ToggleSubscribe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := request.GetProfileId(w, r)
		slug := r.PathValue("slug")

		result, err := h.ChannelService.ToggleSubscribe(slug, userId)

		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		res.Json(w, result, http.StatusOK)
	}
}
