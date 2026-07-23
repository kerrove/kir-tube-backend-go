package res

import (
	"errors"
	"net/http"

	"gorm.io/gorm"
)

func WriteServiceError(w http.ResponseWriter, err, msg error) {
	if errors.Is(err, msg) || errors.Is(err, gorm.ErrRecordNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
