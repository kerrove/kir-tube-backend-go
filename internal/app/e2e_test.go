package app_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"

	"go/kir-tube/internal/app"
	"go/kir-tube/pkg/db"
	"go/kir-tube/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// apiClient drives the assembled app over real HTTP.
type apiClient struct {
	t      *testing.T
	base   string
	client *http.Client
}

func newServer(t *testing.T) (*apiClient, *db.Db, *testutil.FakeStorage) {
	t.Helper()
	database := testutil.RequireDB(t)
	storage := testutil.NewFakeStorage()

	handler, _ := app.Assemble(testutil.TestConfig(), database, storage)
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	return &apiClient{t: t, base: srv.URL, client: &http.Client{Jar: jar}}, database, storage
}

func (c *apiClient) do(method, path, token string, body any) *http.Response {
	c.t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(c.t, err)
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, c.base+path, reader)
	require.NoError(c.t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.client.Do(req)
	require.NoError(c.t, err)
	return resp
}

func decode(t *testing.T, resp *http.Response, target any) {
	t.Helper()
	defer resp.Body.Close()
	require.NoError(t, json.NewDecoder(resp.Body).Decode(target))
}

// register creates a fresh account and returns (userID, accessToken).
func (c *apiClient) register(email, pass string) (string, string) {
	c.t.Helper()
	resp := c.do(http.MethodPost, "/api/auth/register", "", map[string]string{
		"email": email, "password": pass,
	})
	require.Equal(c.t, http.StatusOK, resp.StatusCode)
	var out struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
		AccessToken string `json:"accessToken"`
	}
	decode(c.t, resp, &out)
	require.NotEmpty(c.t, out.AccessToken)
	require.NotEmpty(c.t, out.User.ID)
	return out.User.ID, out.AccessToken
}

func TestE2E_AuthFlow(t *testing.T) {
	c, _, _ := newServer(t)

	// Register.
	_, token := c.register("auth@b.c", "qwerty123")

	// Login with the same credentials.
	resp := c.do(http.MethodPost, "/api/auth/login", "", map[string]string{
		"email": "auth@b.c", "password": "qwerty123",
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Refresh cookie now set on the jar → access-token succeeds.
	resp = c.do(http.MethodPost, "/api/auth/access-token", "", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Wrong password is rejected.
	resp = c.do(http.MethodPost, "/api/auth/login", "", map[string]string{
		"email": "auth@b.c", "password": "wrong",
	})
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// Token works against a protected endpoint.
	resp = c.do(http.MethodGet, "/api/users/profile", token, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestE2E_ProfileIncludesChannelAndToken(t *testing.T) {
	c, database, _ := newServer(t)

	userID, token := c.register("profile@b.c", "qwerty123")

	// Unauthenticated request is rejected.
	resp := c.do(http.MethodGet, "/api/users/profile", "", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// Without a channel yet.
	resp = c.do(http.MethodGet, "/api/users/profile", token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var before struct {
		Email             string          `json:"email"`
		VerificationToken *string         `json:"verificationToken"`
		Channel           json.RawMessage `json:"channel"`
	}
	decode(t, resp, &before)
	assert.Equal(t, "profile@b.c", before.Email)
	assert.Equal(t, "null", strings.TrimSpace(string(before.Channel)))

	// Give the user a channel, then it appears in the profile.
	testutil.CreateChannel(t, database, userID, "my-channel")

	resp = c.do(http.MethodGet, "/api/users/profile", token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var after struct {
		Channel struct {
			Slug string `json:"slug"`
		} `json:"channel"`
	}
	decode(t, resp, &after)
	assert.Equal(t, "my-channel", after.Channel.Slug)
}

func TestE2E_ChannelFlow(t *testing.T) {
	c, database, _ := newServer(t)

	subID, token := c.register("subscriber@b.c", "qwerty123")
	_ = subID

	owner := testutil.CreateUser(t, database, "owner@b.c", "p")
	testutil.CreateChannel(t, database, owner.ID, "redgroup")

	// Public list.
	resp := c.do(http.MethodGet, "/api/channels", "", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// By slug.
	resp = c.do(http.MethodGet, "/api/channels/by-slug/redgroup", "", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var ch struct {
		Slug string `json:"slug"`
	}
	decode(t, resp, &ch)
	assert.Equal(t, "redgroup", ch.Slug)

	// Toggle subscribe (auth required).
	resp = c.do(http.MethodPatch, "/api/channels/toggle-subscribe/redgroup", token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var sub struct {
		IsSubscribed bool `json:"isSubscribed"`
	}
	decode(t, resp, &sub)
	assert.True(t, sub.IsSubscribed)

	// Unauthenticated toggle is rejected.
	resp = c.do(http.MethodPatch, "/api/channels/toggle-subscribe/redgroup", "", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}

func TestE2E_StudioAndVideoFlow(t *testing.T) {
	c, database, _ := newServer(t)

	userID, token := c.register("creator@b.c", "qwerty123")
	testutil.CreateChannel(t, database, userID, "creator-channel")

	// Create a video through the studio (requires a channel).
	resp := c.do(http.MethodPost, "/api/studio/videos", token, map[string]any{
		"title":       "My First Video",
		"description": "hello world",
		"tags":        []string{"golang"},
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var videoID string
	decode(t, resp, &videoID)
	require.NotEmpty(t, videoID)

	// It shows up in the studio listing.
	resp = c.do(http.MethodGet, "/api/studio/videos", token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var listing struct {
		Videos     []struct{ ID, PublicId, Title string } `json:"videos"`
		TotalCount int64                                  `json:"totalCount"`
	}
	decode(t, resp, &listing)
	require.GreaterOrEqual(t, len(listing.Videos), 1)
	publicID := listing.Videos[0].PublicId

	// Public video read by publicId.
	resp = c.do(http.MethodGet, "/api/videos/by-publicId/"+publicID, "", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Increment views count.
	resp = c.do(http.MethodPut, "/api/videos/update-views-count/"+publicID, "", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Update the video.
	newTitle := "Renamed"
	resp = c.do(http.MethodPut, "/api/studio/videos/"+videoID, token, map[string]any{"title": newTitle})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Studio create without a channel is forbidden.
	_, otherToken := c.register("nochannel@b.c", "qwerty123")
	resp = c.do(http.MethodPost, "/api/studio/videos", otherToken, map[string]any{"title": "x"})
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()

	// Delete the video.
	resp = c.do(http.MethodDelete, "/api/studio/videos/"+videoID, token, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Deleting again → 404.
	resp = c.do(http.MethodDelete, "/api/studio/videos/"+videoID, token, nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestE2E_CommentFlow(t *testing.T) {
	c, database, _ := newServer(t)

	userID, token := c.register("commenter@b.c", "qwerty123")
	ch := testutil.CreateChannel(t, database, userID, "c-channel")
	v := testutil.CreateVideo(t, database, ch.ID, "Video")

	// Create a comment.
	resp := c.do(http.MethodPost, "/api/comments", token, map[string]any{
		"text":    "great video",
		"videoId": v.ID,
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var created struct {
		ID string `json:"id"`
	}
	decode(t, resp, &created)
	require.NotEmpty(t, created.ID)

	// List comments for the video.
	resp = c.do(http.MethodGet, "/api/comments/by-video/"+v.PublicId, "", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var list []struct {
		Text string `json:"text"`
	}
	decode(t, resp, &list)
	require.Len(t, list, 1)
	assert.Equal(t, "great video", list[0].Text)

	// Update it.
	resp = c.do(http.MethodPut, "/api/comments/"+created.ID, token, map[string]any{"text": "edited"})
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Delete it.
	resp = c.do(http.MethodDelete, "/api/comments/"+created.ID, token, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Unauthenticated create is rejected.
	resp = c.do(http.MethodPost, "/api/comments", "", map[string]any{"text": "x", "videoId": v.ID})
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}

func TestE2E_PlaylistFlow(t *testing.T) {
	c, database, _ := newServer(t)

	userID, token := c.register("curator@b.c", "qwerty123")
	ch := testutil.CreateChannel(t, database, userID, "cur-channel")
	v := testutil.CreateVideo(t, database, ch.ID, "Video")

	// Create a playlist.
	resp := c.do(http.MethodPost, "/api/playlists", token, map[string]any{"title": "Favourites"})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var pl struct {
		ID string `json:"id"`
	}
	decode(t, resp, &pl)
	require.NotEmpty(t, pl.ID)

	// Toggle a video into it.
	resp = c.do(http.MethodPost, "/api/playlists/"+pl.ID+"/toggle-video", token, map[string]any{"videoId": v.ID})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// List the user's playlists.
	resp = c.do(http.MethodGet, "/api/playlists", token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Unauthenticated access is rejected.
	resp = c.do(http.MethodGet, "/api/playlists", "", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}

func TestE2E_MediaUploadAndServe(t *testing.T) {
	c, _, storage := newServer(t)

	_, token := c.register("uploader@b.c", "qwerty123")

	// Upload an image (multipart, field "file").
	resp := c.uploadFile(token, "avatars", "avatar.png", "image/png", []byte("PNGDATA"))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var uploaded []struct {
		Url  string `json:"url"`
		Name string `json:"name"`
	}
	decode(t, resp, &uploaded)
	require.Len(t, uploaded, 1)
	assert.True(t, strings.HasPrefix(uploaded[0].Url, "/uploads/avatars/"))

	key := "avatars/" + uploaded[0].Name
	assert.True(t, storage.Has(key), "object should be stored under %q", key)

	// Serve it back from /uploads (root, outside /api).
	resp = c.do(http.MethodGet, uploaded[0].Url, "", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	served, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Equal(t, "PNGDATA", string(served))

	// The /api/uploads alias also serves (backward compatibility).
	resp = c.do(http.MethodGet, "/api"+uploaded[0].Url, "", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Missing object → 404.
	resp = c.do(http.MethodGet, "/uploads/avatars/missing.png", "", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()

	// Directory listing is forbidden.
	resp = c.do(http.MethodGet, "/uploads", "", nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()

	// Upload without auth is rejected.
	resp = c.uploadFile("", "avatars", "x.png", "image/png", []byte("x"))
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// Invalid folder → 400.
	resp = c.uploadFile(token, "hackers", "x.png", "image/png", []byte("x"))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func (c *apiClient) uploadFile(token, folder, filename, contentType string, data []byte) *http.Response {
	c.t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	hdr := make(map[string][]string)
	hdr["Content-Disposition"] = []string{`form-data; name="file"; filename="` + filename + `"`}
	hdr["Content-Type"] = []string{contentType}
	part, err := mw.CreatePart(hdr)
	require.NoError(c.t, err)
	_, err = part.Write(data)
	require.NoError(c.t, err)
	require.NoError(c.t, mw.Close())

	req, err := http.NewRequest(http.MethodPost, c.base+"/api/upload-file?folder="+folder, &buf)
	require.NoError(c.t, err)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.client.Do(req)
	require.NoError(c.t, err)
	return resp
}
