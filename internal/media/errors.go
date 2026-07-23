package media

import "errors"

// Validation errors returned by the media package. The handler maps each to a
// 400 Bad Request, mirroring the NestJS BadRequestException the pipes raised.
var (
	ErrNoFile          = errors.New("no file provided")
	ErrUnsupportedType = errors.New("unsupported file type")
	ErrFileTooBig      = errors.New("file size is too big")
	ErrInvalidFolder   = errors.New("invalid folder name")
)
