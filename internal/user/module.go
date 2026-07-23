package user

import (
	"go/kir-tube/configs"
	"go/kir-tube/pkg/db"
	"go/kir-tube/pkg/di"
	"net/http"
)

type UserModuleDeps struct {
	Config          *configs.Config
	Db              *db.Db
	VideoRepository di.IVideoRepository
	UserProvider    di.IUserProvider
}
type UserModule struct {
	UserService *UserService
}

func NewUserModule(router *http.ServeMux, deps UserModuleDeps) *UserModule {
	userRepository := NewUserRepository(deps.Db)
	userService :=
		NewUserService(&UserServiceDeps{
			UserRepository:  userRepository,
			VideoRepository: deps.VideoRepository,
		})
	NewUserHandler(router, UserHandlerDeps{
		UserService:  userService,
		Config:       deps.Config,
		UserProvider: deps.UserProvider,
	})

	return &UserModule{UserService: userService}
}
