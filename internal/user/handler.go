package user

import (
	"go/kir-tube/configs"
	"go/kir-tube/pkg/logs"
	"go/kir-tube/pkg/middleware"
	"go/kir-tube/pkg/res"
	"net/http"
)

type ManagerResponse struct {
	Text string `json:"text"`
}

type UserHandlerDeps struct {
	UserService *UserService
	Config      *configs.Config
}
type UserHandler struct {
	UserService *UserService
	Config      *configs.Config
}

func NewUserHandler(router *http.ServeMux, deps UserHandlerDeps) {
	handler := &UserHandler{
		UserService: deps.UserService,
		Config:      deps.Config,
	}

	logs.RouteLog(router, "GET /users/profile", middleware.IsAuthed(handler.GetMyProfile(), deps.Config))

}

func (h *UserHandler) GetAll() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		users := h.UserService.GetAll()
		res.Json(w, users, 200)
	}
}
func (h *UserHandler) GetMyProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		id, ok := req.Context().Value(middleware.ContextIdKey).(string)
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		users, err := h.UserService.GetById(id)

		if err != nil {
			http.Error(w, ErrUserNotExist, http.StatusNotFound)
			return
		}

		res.Json(w, users, 200)
	}
}

func (h *UserHandler) GetManager() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		res.Json(w, &ManagerResponse{Text: "Manager content"}, 200)
	}
}
func (h *UserHandler) GetPremium() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		res.Json(w, &ManagerResponse{Text: "Premium content"}, 200)
	}
}
