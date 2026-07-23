package media

// Resolution is a named target size the transcoder can output. Mirrors the
// IResolution / RESOLUTIONS table from the NestJS module.
type Resolution struct {
	Name   string
	Width  int
	Height int
}

// Resolutions lists the transcode targets from highest to lowest. processVideo
// keeps only those that do not exceed the source resolution.
var Resolutions = []Resolution{
	{Name: Quality4K, Width: 3840, Height: 2160},
	{Name: Quality2K, Width: 2560, Height: 1440},
	{Name: Quality1080, Width: 1920, Height: 1080},
	{Name: Quality720, Width: 1280, Height: 720},
	{Name: Quality480, Width: 854, Height: 480},
	{Name: Quality360, Width: 640, Height: 360},
}
