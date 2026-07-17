package request

import (
	"go/kir-tube/pkg/middleware"
	"net/http"
)

func GetProfileId(w http.ResponseWriter, r *http.Request) string {
	id, ok := r.Context().Value(middleware.ContextIdKey).(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return ""
	}
	return id
}
