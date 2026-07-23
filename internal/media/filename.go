package media

import (
	"crypto/rand"
	"encoding/hex"
	"path/filepath"
)

// GenerateFilename builds a short, collision-resistant name that keeps the
// original extension, e.g. "1a2b3c4d5e6f.mp4". It mirrors the NestJS
// generateFilename, which concatenated the first two groups of a UUIDv4 (12 hex
// chars) with the extension.
//
// The original's latin1→utf8 re-decode is dropped: it worked around a
// Node/Multer filename-encoding quirk that Go's mime/multipart does not have.
func GenerateFilename(originalName string) string {
	ext := filepath.Ext(originalName)

	buf := make([]byte, 6) // 6 bytes -> 12 hex chars
	if _, err := rand.Read(buf); err != nil {
		panic("media: failed to generate filename: " + err.Error())
	}

	return hex.EncodeToString(buf) + ext
}
