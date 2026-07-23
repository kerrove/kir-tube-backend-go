package middleware

import (
	"context"
	"go/kir-tube/configs"
	"go/kir-tube/pkg/di"
	"go/kir-tube/pkg/jwt"

	"net/http"
	"strings"
)

type key string

const (
	ContextUserKey   key = "ContextUserKey"
	ContextRightsKey key = "ContextRightsKey"
)

func writeUnauthed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
}

func writeForbidden(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(http.StatusText(http.StatusForbidden)))
}

func IsAuthed(next http.Handler, config *configs.Config, users di.IUserProvider) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authedHeader := r.Header.Get("Authorization")

		if !strings.HasPrefix(authedHeader, "Bearer") {
			writeUnauthed(w)
			return
		}

		token := strings.TrimPrefix(authedHeader, "Bearer ")
		isValid, data := jwt.NewJWT(config.Auth.Secret).Parse(token)

		if !isValid {
			writeUnauthed(w)
			return
		}

		// The token still carries only the id; load the full user so the whole
		// record — not just the id — is available in the request context.
		user, err := users.FindContextUser(data.Id)
		if err != nil {
			writeUnauthed(w)
			return
		}

		ctx := context.WithValue(r.Context(), ContextUserKey, user)
		// ctx = context.WithValue(ctx, ContextRightsKey)
		req := r.WithContext(ctx)
		next.ServeHTTP(w, req)
	})
}

// MaybeAuthed is an optional-auth variant of IsAuthed: when a valid Bearer token
// is present it loads the user and puts it into the request context, otherwise
// it lets the request through anonymously (no 401). Handlers behind it must read
// the caller with request.GetProfileIdOptional / GetProfileUserOptional and
// tolerate an empty value.
func MaybeAuthed(next http.Handler, config *configs.Config, users di.IUserProvider) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authedHeader := r.Header.Get("Authorization")

		if strings.HasPrefix(authedHeader, "Bearer") {
			token := strings.TrimPrefix(authedHeader, "Bearer ")
			if isValid, data := jwt.NewJWT(config.Auth.Secret).Parse(token); isValid {
				if user, err := users.FindContextUser(data.Id); err == nil {
					ctx := context.WithValue(r.Context(), ContextUserKey, user)
					r = r.WithContext(ctx)
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}
