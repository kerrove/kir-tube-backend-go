package video

import (
	"crypto/rand"
	"math"
	"net/http"
	"strconv"
)

// publicIDAlphabet is the character set for generated public ids (base36).
const publicIDAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

// publicIDLength mirrors the length of the NestJS backend's nanoid(10) ids so
// migrated data keeps the same public-id shape.
const publicIDLength = 10

// NewPublicID returns a short, URL-safe identifier used in video watch URLs. It
// is distinct from the primary key: the pk stays internal, the public id is the
// value exposed to clients.
func NewPublicID() (string, error) {
	buf := make([]byte, publicIDLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	for i, b := range buf {
		buf[i] = publicIDAlphabet[int(b)%len(publicIDAlphabet)]
	}
	return string(buf), nil
}

func normalizePage(page, limit, defaultLimit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultLimit
	}
	return page, limit
}

func pagesCeil(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(limit)))
}

func unique(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func videoIDs(videos []Video) []string {
	ids := make([]string, len(videos))
	for i, v := range videos {
		ids[i] = v.ID
	}
	return ids
}
func PaginationParams(r *http.Request) (page, limit int) {
	return queryInt(r, "page", 1), queryInt(r, "limit", 10)
}

func queryInt(r *http.Request, key string, fallback int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return fallback
	}
	return value
}
