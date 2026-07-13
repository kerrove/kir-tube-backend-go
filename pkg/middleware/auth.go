package middleware

import (
	"context"
	"go/kir-tube/configs"
	"go/kir-tube/pkg/jwt"

	"net/http"
	"strings"
)

type key string

const (
	ContextIdKey     key = "ContextIdKey"
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

func IsAuthed(next http.Handler, config *configs.Config) http.Handler {
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

		ctx := context.WithValue(r.Context(), ContextIdKey, data.Id)
		// ctx = context.WithValue(ctx, ContextRightsKey)
		req := r.WithContext(ctx)
		next.ServeHTTP(w, req)
	})
}
