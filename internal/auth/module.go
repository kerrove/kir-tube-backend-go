package auth

import (
	"go/kir-tube/configs"
	"go/kir-tube/internal/user"
	"go/kir-tube/pkg/db"
	"net/http"
)

type AuthModuleDeps struct {
	Config      *configs.Config
	Db          *db.Db
	UserService *user.UserService
}

func NewAuthModule(router *http.ServeMux, deps AuthModuleDeps) {
	authService :=
		NewAuthService(&AuthServiceDeps{UserService: deps.UserService, Config: deps.Config})
	NewAuthHandler(router, AuthHandlerDeps{
		AuthService: authService,
		Config:      deps.Config,
	})

}
