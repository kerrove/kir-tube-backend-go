package studio

import (
	"go/kir-tube/configs"
	"go/kir-tube/pkg/db"
	"go/kir-tube/pkg/di"
	"net/http"
)

type StudioModuleDeps struct {
	Config          *configs.Config
	Db              *db.Db
	VideoRepository VideoRepository
	UserProvider    di.IUserProvider
}
type StudioModule struct {
	StudioService *StudioService
}

func NewStudioModule(router *http.ServeMux, deps StudioModuleDeps) *StudioModule {
	studioService :=
		NewStudioService(&StudioServiceDeps{
			VideoRepository: deps.VideoRepository,
		})
	NewStudioHandler(router, StudioHandlerDeps{
		StudioService: studioService,
		Config:        deps.Config,
		UserProvider:  deps.UserProvider,
	})

	return &StudioModule{StudioService: studioService}
}
