package user

import (
	"go/kir-tube/configs"
	"go/kir-tube/pkg/di"
	"go/kir-tube/pkg/logs"
	"go/kir-tube/pkg/middleware"
	request "go/kir-tube/pkg/req"
	"go/kir-tube/pkg/res"
	"net/http"
)

type UserHandlerDeps struct {
	UserService  *UserService
	Config       *configs.Config
	UserProvider di.IUserProvider
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

	logs.RouteLog(router, "GET /users/profile", middleware.IsAuthed(handler.GetMyProfile(), deps.Config, deps.UserProvider))
	logs.RouteLog(router, "PUT /users/profile", middleware.IsAuthed(handler.UpdateMyProfile(), deps.Config, deps.UserProvider))
}

func (h *UserHandler) GetMyProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := request.GetProfileId(w, r)

		users, err := h.UserService.GetProfile(id)

		if err != nil {
			http.Error(w, ErrUserNotExist, http.StatusNotFound)
			return
		}

		res.Json(w, users, http.StatusOK)
	}
}

func (h *UserHandler) UpdateMyProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := request.GetProfileId(w, r)
		if id == "" {
			return
		}

		body, err := request.HandleBody[UpdateProfileReq](&w, r)
		if err != nil {
			return
		}

		profile, err := h.UserService.UpdateProfile(id, body)
		if err != nil {
			http.Error(w, ErrUserNotExist, http.StatusNotFound)
			return
		}

		res.Json(w, profile, http.StatusOK)
	}
}
