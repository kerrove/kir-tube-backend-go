package auth

import (
	"go/kir-tube/configs"
	"go/kir-tube/pkg/logs"
	request "go/kir-tube/pkg/req"
	"go/kir-tube/pkg/res"
	"net/http"
)

const (
	RefreshTokenName      = "refreshToken"
	ExpireDayRefreshToken = 1
	Domain                = "localhost"
)

type AuthHandlerDeps struct {
	*configs.Config
	*AuthService
}
type AuthHandler struct {
	*AuthService
	*configs.Config
}

func NewAuthHandler(router *http.ServeMux, deps AuthHandlerDeps) {
	handler := &AuthHandler{
		Config:      deps.Config,
		AuthService: deps.AuthService,
	}

	logs.RouteLog(router, "POST /auth/access-token", handler.GetNewTokens())
	logs.RouteLog(router, "POST /auth/register", handler.Register())
	logs.RouteLog(router, "POST /auth/logout", handler.Logout())
	logs.RouteLog(router, "POST /auth/login", handler.Login())
	logs.RouteLog(router, "POST /verify-email", handler.VerifyEmail())
}

func (handler *AuthHandler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		body, err := request.HandleBody[AuthRequest](&w, req)
		if err != nil {
			return
		}

		user, err := handler.AuthService.Login(*body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		handler.AuthService.addRefreshTokenToResponse(w, user.RefreshToken)
		res.Json(w, user, 200)
	}
}
func (handler *AuthHandler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		body, err := request.HandleBody[AuthRequest](&w, req)
		if err != nil {
			return
		}

		user, err := handler.AuthService.Register(*body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		handler.AuthService.addRefreshTokenToResponse(w, user.RefreshToken)
		res.Json(w, user, 200)
	}
}

func (handler *AuthHandler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		handler.AuthService.RemoveRefreshTokenFromResponse(w)

		res.Json(w, true, 200)
	}

}

func (h *AuthHandler) VerifyEmail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")

		if token == "" {
			http.Error(w, "Token not passed", http.StatusUnauthorized)
			return
		}

		h.AuthService.verifyEmail(token)

		res.Json(w, true, 200)
	}

}
func (handler *AuthHandler) GetNewTokens() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		cookie, err := req.Cookie(RefreshTokenName)
		if err != nil {
			handler.AuthService.RemoveRefreshTokenFromResponse(w)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		resp, err := handler.AuthService.GetNewTokens(cookie.Value)
		if err != nil {
			handler.AuthService.RemoveRefreshTokenFromResponse(w)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		handler.AuthService.addRefreshTokenToResponse(w, resp.RefreshToken)

		res.Json(w, resp, 200)
	}
}
