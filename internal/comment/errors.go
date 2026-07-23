package comment

import "errors"

const (
	ErrCommentNotExist = "Комментарий не найден"
	ErrCommentCantEdit = "Вы не можете редактировать этот комментарий"
)

var ErrCommentForbidden = errors.New(ErrCommentCantEdit)
