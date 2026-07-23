package media

import "strings"

// allowedMimeTypes are the content types the upload endpoint accepts, matching
// the NestJS FileValidationPipe.
var allowedMimeTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/svg+xml":   true,
	"video/mp4":       true,
	"video/quicktime": true,
}

// allowedFolders are the destination folders the upload endpoint accepts,
// matching the NestJS FolderValidationPipe.
var allowedFolders = map[string]bool{
	"avatars":    true,
	"banners":    true,
	"thumbnails": true,
	"videos":     true,
}

// maxFileSize is the per-file upload cap (300 MB), matching the pipe.
const maxFileSize = 300 * 1024 * 1024

// ValidateFiles enforces the FileValidationPipe rules: at least one file, each
// with an allowed mime type and within the size cap.
func ValidateFiles(files []UploadFile) error {
	if len(files) == 0 {
		return ErrNoFile
	}

	for _, file := range files {
		if file.MimeType == "" {
			return ErrNoFile
		}
		if !allowedMimeTypes[file.MimeType] {
			return ErrUnsupportedType
		}
		if file.Size > maxFileSize {
			return ErrFileTooBig
		}
	}

	return nil
}

// ValidateFolder enforces the FolderValidationPipe rule: an empty folder is
// allowed (the service falls back to "default"), otherwise it must be one of the
// allowed folders (case-insensitive).
func ValidateFolder(folder string) error {
	if folder == "" {
		return nil
	}
	if !allowedFolders[strings.ToLower(folder)] {
		return ErrInvalidFolder
	}
	return nil
}
