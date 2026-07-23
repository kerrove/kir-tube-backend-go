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
	ContextUserKey key = "ContextUserKey"
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

		user, err := users.FindContextUser(data.Id)
		if err != nil {
			writeUnauthed(w)
			return
		}

		ctx := context.WithValue(r.Context(), ContextUserKey, user)
		req := r.WithContext(ctx)
		next.ServeHTTP(w, req)
	})
}

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
