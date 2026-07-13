package user

import (
	"go/kir-tube/configs"
	"go/kir-tube/pkg/db"
	"net/http"
)

type UserModuleDeps struct {
	Config *configs.Config
	Db     *db.Db
}
type UserModule struct {
	UserService *UserService
}

func NewUserModule(router *http.ServeMux, deps UserModuleDeps) *UserModule {
	userRepository := NewUserRepository(deps.Db)
	userService :=
		NewUserService(&UserServiceDeps{UserRepository: userRepository})
	NewUserHandler(router, UserHandlerDeps{
		UserService: userService,
		Config:      deps.Config,
	})

	return &UserModule{UserService: userService}
}
