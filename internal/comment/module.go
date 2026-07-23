package comment

import (
	"go/kir-tube/configs"
	"go/kir-tube/pkg/db"
	"go/kir-tube/pkg/di"
	"net/http"
)

type CommentModuleDeps struct {
	Config          *configs.Config
	Db              *db.Db
	VideoRepository di.IVideoRepository
	UserProvider    di.IUserProvider
}

func NewCommentModule(router *http.ServeMux, deps CommentModuleDeps) {
	commentRepository := NewCommentRepository(deps.Db)

	commentService :=
		NewCommentService(&CommentServiceDeps{VideoRepository: deps.VideoRepository, CommentRepository: commentRepository})
	NewCommentHandler(router, CommentHandlerDeps{
		CommentService: commentService,
		Config:         deps.Config,
		UserProvider:   deps.UserProvider,
	})

}
