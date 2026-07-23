package media

// Player-quality labels, mirroring the NestJS EnumVideoPlayerQuality. They are
// plain strings because video.Video.MaxResolution is a string column
// (defaulting to "1080p").
const (
	Quality4K   = "4K"
	Quality2K   = "2K"
	Quality1080 = "1080p"
	Quality720  = "720p"
	Quality480  = "480p"
	Quality360  = "360p"
)

// UploadFile is one decoded multipart file, the Go counterpart of the NestJS
// IFile (Multer memory-storage file). The whole payload is held in Buffer,
// matching the original memory-storage behaviour.
type UploadFile struct {
	Name         string
	OriginalName string
	MimeType     string
	Size         int64
	Buffer       []byte
}

// MediaResponse is one saved media descriptor returned to the client. It mirrors
// IMediaResponse: MaxResolution is only set for videos.
type MediaResponse struct {
	Url           string `json:"url"`
	Name          string `json:"name"`
	MaxResolution string `json:"maxResolution,omitempty"`
}
