package request

import (
	"go/kir-tube/pkg/di"
	"go/kir-tube/pkg/middleware"
	"net/http"
)

// GetProfileUser returns the authenticated user placed in the context by
// middleware.IsAuthed. If the request did not pass through auth (no user in
// context) it writes a 401 and returns nil.
func GetProfileUser(w http.ResponseWriter, r *http.Request) *di.ContextUser {
	user, ok := r.Context().Value(middleware.ContextUserKey).(*di.ContextUser)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return nil
	}
	return user
}

// GetProfileUserOptional returns the authenticated user if the request carries
// one (i.e. it passed through MaybeAuthed with a valid token), or nil for
// anonymous callers. Unlike GetProfileUser it never writes a 401.
func GetProfileUserOptional(r *http.Request) *di.ContextUser {
	user, _ := r.Context().Value(middleware.ContextUserKey).(*di.ContextUser)
	return user
}

func GetProfileId(w http.ResponseWriter, r *http.Request) string {
	user := GetProfileUser(w, r)
	if user == nil {
		return ""
	}
	return user.ID
}

// GetProfileIdOptional returns the profile id if the request carries a user
// (i.e. it passed through MaybeAuthed with a valid token), or "" for anonymous
// callers. Unlike GetProfileId it never writes a 401.
func GetProfileIdOptional(r *http.Request) string {
	if user := GetProfileUserOptional(r); user != nil {
		return user.ID
	}
	return ""
}
