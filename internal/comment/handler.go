package comment

import (
	"errors"
	"go/kir-tube/configs"
	"go/kir-tube/internal/video"
	"go/kir-tube/pkg/di"
	"go/kir-tube/pkg/logs"
	"go/kir-tube/pkg/middleware"
	request "go/kir-tube/pkg/req"
	"go/kir-tube/pkg/res"
	"net/http"
)

type CommentHandlerDeps struct {
	CommentService *CommentService
	Config         *configs.Config
	UserProvider   di.IUserProvider
}
type CommentHandler struct {
	CommentService *CommentService
	Config         *configs.Config
}

func NewCommentHandler(router *http.ServeMux, deps CommentHandlerDeps) {
	handler := &CommentHandler{
		CommentService: deps.CommentService,
		Config:         deps.Config,
	}

	logs.RouteLog(router, "GET /comments/by-video/{publicId}", handler.GetByPublicId())

	logs.RouteLog(router, "POST /comments", middleware.IsAuthed(handler.CreateComment(), deps.Config, deps.UserProvider))
	logs.RouteLog(router, "PUT /comments/{publicId}", middleware.IsAuthed(handler.UpdateComment(), deps.Config, deps.UserProvider))
	logs.RouteLog(router, "DELETE /comments/{publicId}", middleware.IsAuthed(handler.DeleteComment(), deps.Config, deps.UserProvider))
}

func (h *CommentHandler) GetByPublicId() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		publicId := r.PathValue("publicId")

		videos, err := h.CommentService.GetByVideo(publicId)
		if err != nil {
			res.WriteServiceError(w, err, video.ErrVideoNotFound)
		}

		res.Json(w, videos, http.StatusOK)
	}
}
func (h *CommentHandler) CreateComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := request.GetProfileId(w, r)

		body, err := request.HandleBody[CreateCommentReq](&w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		comment, err := h.CommentService.CreateComment(userId, body)
		if err != nil {
			res.WriteServiceError(w, err, errors.New(ErrCommentNotExist))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		res.Json(w, comment, http.StatusOK)
	}
}
func (h *CommentHandler) DeleteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := request.GetProfileId(w, r)
		publicId := r.PathValue("publicId")

		comment, err := h.CommentService.DeleteComment(publicId, userId)
		if err != nil {
			res.WriteServiceError(w, err, errors.New(ErrCommentNotExist))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		res.Json(w, comment, http.StatusOK)
	}
}
func (h *CommentHandler) UpdateComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := request.GetProfileId(w, r)
		publicId := r.PathValue("publicId")

		body, err := request.HandleBody[UpdateCommentReq](&w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		comment, err := h.CommentService.UpdateComment(publicId, userId, body)
		if err != nil {
			res.WriteServiceError(w, err, errors.New(ErrCommentNotExist))
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		res.Json(w, comment, http.StatusOK)
	}
}
